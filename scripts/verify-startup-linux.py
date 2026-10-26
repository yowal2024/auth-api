"""Verify the release and optional original binary with disposable local services.

Run on Linux with Docker, for example from WSL. No production configuration,
credentials, SQL dumps, or existing containers are used.
"""
import argparse
from datetime import datetime, timezone
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import socket
import ssl
import subprocess
import tempfile
import time
import urllib.error
import urllib.request
import uuid


def service_environment():
    # Configuration overrides from the caller must not redirect tests to another database.
    return {key: value for key, value in os.environ.items() if not key.startswith("AUTH_API_")}


def docker(*args):
    return subprocess.check_output(["docker", *args], text=True, stderr=subprocess.PIPE).strip()


def free_port():
    with socket.socket() as connection:
        connection.bind(("127.0.0.1", 0))
        return connection.getsockname()[1]


def listen_ports(process):
    inodes = set()
    for fd in Path(f"/proc/{process.pid}/fd").iterdir():
        try:
            target = os.readlink(fd)
        except FileNotFoundError:
            continue
        if target.startswith("socket:["):
            inodes.add(target[8:-1])
    ports = set()
    for name in ("tcp", "tcp6"):
        for line in Path(f"/proc/{process.pid}/net/{name}").read_text().splitlines()[1:]:
            fields = line.split()
            if fields[3] == "0A" and fields[9] in inodes:
                ports.add(int(fields[1].rsplit(":", 1)[1], 16))
    return sorted(ports)


def request(base, path, body=None, token="", context=None, raw=False, expected_code=0):
    headers = {}
    if token:
        headers["Authorization"] = "Bearer " + token
    encoded = None
    if body is not None:
        encoded = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    opener = urllib.request.build_opener(
        urllib.request.ProxyHandler({}), urllib.request.HTTPSHandler(context=context)
    )
    req = urllib.request.Request(base + path, data=encoded, headers=headers)
    with opener.open(req, timeout=10) as response:
        result = response.read()
    if raw:
        return result
    result = json.loads(result)
    if result.get("code") != expected_code:
        raise AssertionError(f"{path}: code={result.get('code')} message={result.get('message', result.get('msg'))}")
    return result.get("data")


class Service:
    def __init__(self, executable, cwd, args, log):
        self.log = log
        self.output = log.open("wb")
        self.process = subprocess.Popen(
            [str(executable), *args], cwd=cwd, env=service_environment(),
            stdout=self.output, stderr=subprocess.STDOUT
        )

    def ready(self, base, context=None):
        deadline = time.monotonic() + 25
        while time.monotonic() < deadline:
            if self.process.poll() is not None:
                raise AssertionError("service exited: " + self.log.read_text())
            try:
                return request(base, "/api/install/status", context=context)
            except (OSError, ValueError):
                time.sleep(0.1)
        raise AssertionError("service did not become ready: " + self.log.read_text())

    def stop(self):
        if self.process.poll() is None:
            self.process.terminate()
            try:
                self.process.wait(timeout=12)
            except subprocess.TimeoutExpired:
                self.process.kill()
                self.process.wait()
        self.output.close()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", required=True, type=Path)
    parser.add_argument("--original", type=Path)
    parser.add_argument("--server-tests", type=Path, help="optional compiled Linux server test binary")
    parser.add_argument("--report", type=Path)
    args = parser.parse_args()
    artifact = args.binary.resolve()
    original = args.original.resolve() if args.original else None
    report = {"checks": [], "sha256": hashlib.sha256(artifact.read_bytes()).hexdigest()}
    if original:
        report["original_sha256"] = hashlib.sha256(original.read_bytes()).hexdigest()
    suffix = uuid.uuid4().hex[:12]
    mysql_name = "auth-startup-mysql-" + suffix
    redis_name = "auth-startup-redis-" + suffix
    database_password = "test_" + uuid.uuid4().hex
    containers = []
    services = []

    def passed(name):
        report["checks"].append(name)
        print("PASS " + name, flush=True)

    try:
        docker("run", "--rm", "-d", "--name", mysql_name,
               "-p", "127.0.0.1::3306", "-e", "MYSQL_ROOT_PASSWORD=" + database_password,
               "-e", "MYSQL_ROOT_HOST=%", "-e", "MYSQL_DATABASE=auth", "mysql:8.4",
               "--skip-log-bin", "--innodb-buffer-pool-size=64M", "--innodb-redo-log-capacity=16M")
        containers.append(mysql_name)
        docker("run", "--rm", "-d", "--name", redis_name,
               "-p", "127.0.0.1::6379", "redis:7.4-alpine",
               "redis-server", "--save", "", "--appendonly", "no")
        containers.append(redis_name)
        mysql_port = int(docker("port", mysql_name, "3306/tcp").rsplit(":", 1)[1])
        redis_port = int(docker("port", redis_name, "6379/tcp").rsplit(":", 1)[1])

        def sql(statement):
            return docker("exec", "-e", "MYSQL_PWD=" + database_password, mysql_name,
                          "mysql", "-h", "127.0.0.1", "-u", "root", "--connect-timeout=2", "--batch", "--skip-column-names",
                          "auth", "-e", statement)

        def redis_command(*arguments):
            return docker("exec", redis_name, "redis-cli", "--raw", *arguments)

        last_error = ""
        for attempt in range(240):
            try:
                sql("SELECT 1")
                break
            except subprocess.CalledProcessError as error:
                last_error = error.stderr.replace(database_password, "[test password]")
                time.sleep(0.5)
        else:
            logs = subprocess.run(["docker", "logs", "--tail", "40", mysql_name], capture_output=True, text=True)
            raise AssertionError("isolated MySQL did not start:\n" + last_error + logs.stdout + logs.stderr)
        passed("isolated MySQL and Redis ready")

        with tempfile.TemporaryDirectory(prefix="auth-startup-") as temporary:
            root = Path(temporary)
            app = root / "service"
            configs = app / "configs"
            configs.mkdir(parents=True)
            caller = root / "unrelated-working-directory"
            caller.mkdir()
            stray = caller / "config.yaml"
            stray.write_text("installed: false\nserver:\n  addr: '127.0.0.1:1'\n")
            stray_bytes = stray.read_bytes()
            binary = app / "auth-api-linux-amd64"
            shutil.copyfile(artifact, binary)
            binary.chmod(0o700)
            config = configs / "config.yaml"
            config.write_text(
                "installed: false\nserver:\n  addr: ':8080'\n  mode: prod\n"
                f"mysql:\n  dsn: root:{database_password}@tcp(127.0.0.1:{mysql_port})/auth?parseTime=true&loc=Local&charset=utf8mb4\n"
                "  max_open: 50\n  max_idle: 10\n  conn_lifetime: 3600\n  table_prefix: a_\n"
                f"redis:\n  addr: '127.0.0.1:{redis_port}'\n  db: 0\n  password: ''\n"
                "clone:\n  canonical_dir: ./uploads/canonical\n  output_dir: ./uploads/diffs\n  diff_ttl_sec: 60\n"
                "upload:\n  root: ./uploads\n  url_prefix: /uploads\n"
                "session:\n  ttl_sec: 604800\nlog:\n  level: info\nweb:\n  dist: ''\n"
            )
            config.chmod(0o600)
            port = free_port()
            plain = f"http://127.0.0.1:{port}"
            secure = f"https://127.0.0.1:{port}"

            def start(executable=binary, explicit=False):
                flags = ["-port", str(port)]
                if explicit:
                    flags += ["-config", str(config)]
                service = Service(executable, app if explicit else caller, flags,
                                  root / f"service-{len(services)}.log")
                services.append(service)
                return service

            service = start()
            assert service.ready(plain)["installed"] is False
            assert Path(f"/proc/{service.process.pid}/cwd").resolve() == app
            assert listen_ports(service.process) == [port]
            assert str(config) in service.log.read_text()
            assert not (app / "data/master.key").exists()
            passed("executable config wins over unrelated working-directory config")
            assert sql("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE()") == "0"
            passed("uninstalled startup leaves the database untouched until installer migration")

            mysql = {"host": "127.0.0.1", "port": mysql_port, "user": "root",
                     "password": database_password, "database": "auth", "tablePrefix": "a_"}
            redis = {"addr": f"127.0.0.1:{redis_port}", "db": 0, "password": ""}
            request(plain, "/api/install/save-config", {"mysql": mysql, "redis": redis})
            migrated = request(plain, "/api/install/migrate", {})
            assert len(migrated["applied"]) == 46
            admin_password = "Test-" + uuid.uuid4().hex[:10]
            request(plain, "/api/install/create-admin",
                    {"username": "admin", "password": admin_password, "adminEntry": "admin"})
            finalized = request(plain, "/api/install/finalize", {})
            assert finalized["restartRequired"] is False
            assert request(plain, "/api/install/status")["installed"] is True
            assert re.search(r"addr:\s*['\"]?:8080", config.read_text())
            passed("installation activates backend on the same port without persisting -port")

            token = request(plain, "/api/admin/login",
                            {"username": "admin", "password": admin_password})["token"]
            certificate = request(plain, "/api/admin/system/https/gen-cert",
                                  {"host": "127.0.0.1,localhost"}, token)["cert"]
            context = ssl.create_default_context(cadata=certificate)
            context.minimum_version = ssl.TLSVersion.TLSv1_2
            assert (app / "data/master.key").exists()
            result = request(plain, "/api/admin/system/https", {"enabled": True}, token)
            assert result["hotSwap"] is True
            assert request(secure, "/api/install/status", context=context)["installed"] is True
            assert listen_ports(service.process) == [port]
            assert sql("SELECT https_addr FROM a_settings WHERE id=1") == ":8443"
            try:
                request(plain, "/api/install/status")
                raise AssertionError("HTTP remained active after HTTPS was enabled")
            except urllib.error.HTTPError as error:
                assert error.code == 400
            passed("HTTPS uses only -port, ignoring legacy https_addr :8443")

            html = request(secure, "/admin/login", context=context, raw=True).decode()
            asset = re.search(r'src="(/assets/[^\"]+\.js)"', html).group(1)
            entry_js = request(secure, asset, context=context, raw=True).decode()
            assert len(entry_js) > 1000
            assert not (app / "webdist").exists()
            passed("single binary serves embedded frontend over HTTPS")
            build_info = json.loads(request(secure, "/build-info.json", context=context, raw=True))
            assert re.fullmatch(r"\d+\.\d+\.\d+", build_info["version"])
            assert datetime.fromisoformat(build_info["builtAt"].replace("Z", "+00:00")) <= datetime.now(timezone.utc)
            layout_asset = re.search(r'assets/AdminLayout-[^"\s]+\.js', entry_js).group(0)
            layout_js = request(secure, "/" + layout_asset, context=context, raw=True).decode()
            assert build_info["version"] in layout_js and build_info["builtAt"] in layout_js
            assert "构建信息" in layout_js and "2026-08-03 20:32" not in layout_js
            report["build_info"] = build_info
            passed("embedded footer contains the release version and actual build timestamp")

            result = request(secure, "/api/admin/system/https", {"enabled": False}, token, context)
            assert result["hotSwap"] is True
            assert request(plain, "/api/install/status")["installed"] is True
            assert listen_ports(service.process) == [port]
            request(plain, "/api/admin/system/https", {"enabled": True}, token)
            assert request(secure, "/api/install/status", context=context)["installed"] is True
            service.stop()
            passed("live HTTPS-to-HTTP-to-HTTPS switch completes its save requests")

            config_before = config.read_bytes()
            master_before = (app / "data/master.key").read_bytes()
            service = start()
            assert service.ready(secure, context)["installed"] is True
            assert listen_ports(service.process) == [port]
            assert "database migrations applied: []" in service.log.read_text()
            assert config.read_bytes() == config_before
            assert (app / "data/master.key").read_bytes() == master_before
            assert stray.read_bytes() == stray_bytes
            assert json.loads(request(secure, "/build-info.json", context=context, raw=True)) == build_info
            service.stop()
            passed("restart loads existing TLS, master key and config from the service directory")

            admin_id = int(sql("SELECT id FROM a_admins WHERE username='admin'"))
            admin_index = f"sess:admin:{admin_id}"
            legacy_admin_token = None
            if original:
                reference = app / "original-auth-api"
                shutil.copyfile(original, reference)
                reference.chmod(0o700)
                # Remove only disposable test schema to verify the original startup migration.
                sql("DROP TABLE a_extend_files; DELETE FROM a__migrations WHERE version='0041_extend_files.sql'")
                service = start(reference, explicit=True)
                assert service.ready(secure, context)["installed"] is True
                assert listen_ports(service.process) == [port]
                assert sql("SELECT COUNT(*) FROM a__migrations WHERE version='0041_extend_files.sql' AND applied_at>0") == "1"
                assert sql("SELECT COUNT(*) FROM a_extend_files") == "0"
                legacy_admin_token = request(secure, "/api/admin/login",
                                             {"username": "admin", "password": admin_password}, context=context)["token"]
                assert redis_command("TYPE", admin_index) == "string"
                assert redis_command("GET", admin_index) == legacy_admin_token
                service.stop()
                passed("original binary automatically applies pending migrations on startup")
                passed("original login confirmed: admin session index is a Redis string")
                sql("UPDATE a_settings SET https_enabled=0 WHERE id=1")
                service = start(reference, explicit=True)
                assert service.ready(plain)["installed"] is True
                assert listen_ports(service.process) == [port]
                service.stop()
                sql("UPDATE a_settings SET https_enabled=1 WHERE id=1")
                passed("original binary confirmed: one -port listener, HTTPS or HTTP by database switch")

            if legacy_admin_token is None:
                legacy_admin_token = token
                redis_command("SET", admin_index, legacy_admin_token, "EX", "604800")
            legacy_admin_data = redis_command("GET", "sess:" + legacy_admin_token)
            assert legacy_admin_data

            # Recreate the v0041 schema only in this disposable database, preserving its admin and TLS.
            sql("""
                ALTER TABLE a_users DROP COLUMN last_active_time, MODIFY COLUMN ban_expire INT UNSIGNED NOT NULL DEFAULT 0;
                ALTER TABLE a_kamis DROP COLUMN last_active_time, MODIFY COLUMN ban_expire INT UNSIGNED NOT NULL DEFAULT 0;
                ALTER TABLE a_apps MODIFY COLUMN geo_provider VARCHAR(16) NOT NULL DEFAULT 'tencent';
                ALTER TABLE a_logs DROP INDEX idx_logs_app_type_time, DROP COLUMN appid;
                DROP TABLE a_security_events, a_backup_settings, a_backup_runs, a_app_record_ids, a_app_record_counters;
                DELETE FROM a__migrations WHERE version >= '0042';
                INSERT INTO a_apps (name, geo_provider, geo_key, bind_region) VALUES ('preserved-app', 'tencent', 'test-key', 1);
                INSERT INTO a_users (username, password, vip, ban_expire, last_login_time, ip_list)
                    VALUES ('preserved-user', 'test-password-hash', 1900000000, 1800000000, 1700000000, '[]');
                INSERT INTO a_kamis (card, vip, ban_expire, last_login_time)
                    VALUES ('preserved-card', 1900000001, 1800000001, 1700000001);
            """)
            assert sql("SELECT COUNT(*) FROM a__migrations") == "41"

            def existing_data():
                return sql("""
                    SELECT id, username, password, role FROM a_admins ORDER BY id;
                    SELECT id, username, password, vip, ban_expire, last_login_time FROM a_users ORDER BY id;
                    SELECT id, card, vip, ban_expire, last_login_time FROM a_kamis ORDER BY id;
                    SELECT id, name, geo_provider, geo_key, bind_region FROM a_apps ORDER BY id;
                    SELECT * FROM a_settings ORDER BY id;
                """)

            data_before = existing_data()
            journal_before = sql("SELECT version, applied_at FROM a__migrations ORDER BY version")
            service = start()
            assert service.ready(secure, context)["installed"] is True
            assert listen_ports(service.process) == [port]
            assert "database migrations applied: [0042_signed_ban_expire.sql 0043_default_local_geo.sql 0044_security_backup_activity.sql 0045_admin_log_app_scope.sql 0046_app_record_ids.sql]" in service.log.read_text()
            assert sql("SELECT COUNT(*) FROM a__migrations WHERE applied_at>0") == "46"
            assert sql("SELECT version, applied_at FROM a__migrations WHERE version<'0042' ORDER BY version") == journal_before
            assert existing_data() == data_before
            assert sql("SELECT last_active_time FROM a_users WHERE username='preserved-user'") == "1700000000"
            assert sql("SELECT last_active_time FROM a_kamis WHERE card='preserved-card'") == "1700000001"
            assert sql("SELECT COUNT(*) FROM a_backup_settings") == "1"
            assert sql("SELECT COUNT(*) FROM a_security_events") == "0"
            assert config.read_bytes() == config_before
            assert (app / "data/master.key").read_bytes() == master_before
            upgraded_token = request(secure, "/api/admin/login",
                                     {"username": "admin", "password": admin_password}, context=context)["token"]
            assert request(secure, "/api/admin/me", token=upgraded_token, context=context)["id"] == admin_id
            assert redis_command("TYPE", admin_index) == "string"
            assert redis_command("GET", admin_index) == upgraded_token
            assert redis_command("EXISTS", "sess:" + legacy_admin_token) == "0"
            request(secure, "/api/admin/me", token=legacy_admin_token, context=context, expected_code=401)
            passed("new login replaces the original Redis string session and rejects its old token")

            latest_token = request(secure, "/api/admin/login",
                                   {"username": "admin", "password": admin_password}, context=context)["token"]
            request(secure, "/api/admin/me", token=upgraded_token, context=context, expected_code=401)
            request(secure, "/api/admin/logout", {}, upgraded_token, context, expected_code=401)
            assert request(secure, "/api/admin/me", token=latest_token, context=context)["id"] == admin_id
            request(secure, "/api/admin/login", {"username": "admin", "password": "incorrect-password"},
                    context=context, expected_code=1001)
            assert redis_command("GET", admin_index) == latest_token
            assert request(secure, "/api/admin/me", token=latest_token, context=context)["id"] == admin_id
            assert redis_command("EXISTS", "sess:" + upgraded_token) == "0"
            passed("only the latest admin login works; failed login and stale logout do not evict it")

            # Recreate the previous release's set index in disposable Redis.
            other_set_token = "previous-set-" + uuid.uuid4().hex
            redis_command("SET", "sess:" + other_set_token, redis_command("GET", "sess:" + latest_token), "EX", "604800")
            redis_command("DEL", admin_index)
            redis_command("SADD", admin_index, latest_token, other_set_token)
            request(secure, "/api/admin/me", token=latest_token, context=context, expected_code=401)
            current_token = request(secure, "/api/admin/login",
                                    {"username": "admin", "password": admin_password}, context=context)["token"]
            assert redis_command("TYPE", admin_index) == "string"
            assert redis_command("GET", admin_index) == current_token
            assert redis_command("EXISTS", "sess:" + latest_token, "sess:" + other_set_token) == "0"
            assert request(secure, "/api/admin/me", token=current_token, context=context)["id"] == admin_id
            request(secure, "/api/admin/logout", {}, current_token, context)
            assert redis_command("EXISTS", "sess:" + current_token, admin_index) == "0"
            assert redis_command("EXISTS", "sess:" + upgraded_token) == "0"
            assert redis_command("EXISTS", "sess:" + legacy_admin_token) == "0"
            service.stop()
            passed("-port startup upgrades v0041 to v0046 while preserving existing accounts, cards, settings and TLS")
            passed("previous set indexes become one string token; all older sessions are revoked")

            # A replay of 0044 would reset activity values; retain sentinels across another startup.
            sql("UPDATE a_users SET last_active_time=1700000123 WHERE username='preserved-user'; "
                "UPDATE a_kamis SET last_active_time=1700000456 WHERE card='preserved-card'")
            journal_after = sql("SELECT version, applied_at FROM a__migrations ORDER BY version")
            service = start()
            assert service.ready(secure, context)["installed"] is True
            assert "database migrations applied: []" in service.log.read_text()
            assert sql("SELECT version, applied_at FROM a__migrations ORDER BY version") == journal_after
            assert sql("SELECT last_active_time FROM a_users WHERE username='preserved-user'") == "1700000123"
            assert sql("SELECT last_active_time FROM a_kamis WHERE card='preserved-card'") == "1700000456"
            assert existing_data() == data_before
            service.stop()
            passed("second startup skips completed migrations and preserves newer data")

            migration = subprocess.run([str(binary), "-config", str(config), "-migrate"],
                                       cwd=caller, env=service_environment(), capture_output=True, text=True, timeout=30)
            assert migration.returncode == 0, migration.stderr
            assert "database migrations applied: []" in migration.stderr
            passed("optional -migrate maintenance command still exits after checking migrations")

            master = app / "data/master.key"
            saved_master = app / "data/master.key.saved"
            master.rename(saved_master)
            missing = subprocess.run([str(binary), "-port", str(port)], cwd=caller,
                                     env=service_environment(), capture_output=True, text=True, timeout=25)
            assert missing.returncode != 0
            assert "data/master.key" in missing.stderr
            assert not master.exists()
            saved_master.rename(master)
            passed("missing original master key fails startup without generating a replacement")

            with socket.socket() as occupied:
                occupied.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
                occupied.bind(("0.0.0.0", port))
                occupied.listen()
                conflict = subprocess.run([str(binary), "-port", str(port)], cwd=caller,
                                          env=service_environment(), capture_output=True, text=True, timeout=25)
                assert conflict.returncode != 0
                assert "listen on" in conflict.stderr
            passed("occupied port reports an error and exits nonzero")

            # A real SQL failure leaves an interrupted journal entry and must never start serving.
            sql("DELETE FROM a__migrations WHERE version='0044_security_backup_activity.sql'")
            failed = subprocess.run([str(binary), "-port", str(port)], cwd=caller,
                                    env=service_environment(), capture_output=True, text=True, timeout=25)
            assert failed.returncode != 0
            assert "database migration failed: migration 0044_security_backup_activity.sql" in failed.stderr
            assert "listening on" not in failed.stderr
            assert sql("SELECT applied_at FROM a__migrations WHERE version='0044_security_backup_activity.sql'") == "-1"
            assert existing_data() == data_before
            passed("SQL migration failure stops startup with a nonzero exit before listening")

            sql("UPDATE a_users SET last_active_time=1700000789 WHERE username='preserved-user'")
            interrupted = subprocess.run([str(binary), "-port", str(port)], cwd=caller,
                                         env=service_environment(), capture_output=True, text=True, timeout=25)
            assert interrupted.returncode != 0
            assert "database migration failed:" in interrupted.stderr
            assert "0044_security_backup_activity.sql" in interrupted.stderr
            assert "listening on" not in interrupted.stderr
            assert sql("SELECT last_active_time FROM a_users WHERE username='preserved-user'") == "1700000789"
            assert sql("SELECT applied_at FROM a__migrations WHERE version='0044_security_backup_activity.sql'") == "-1"
            passed("interrupted migration is not replayed automatically on a later startup")

            if args.server_tests:
                environment = dict(
                    os.environ,
                    AUTH_RECOVERY_MYSQL_DSN=f"root:{database_password}@tcp(127.0.0.1:{mysql_port})/auth?parseTime=true",
                    AUTH_RECOVERY_REDIS_ADDR=f"127.0.0.1:{redis_port}",
                    AUTH_RECOVERY_REDIS_PASSWORD="",
                    AUTH_RECOVERY_REDIS_DISPOSABLE="1",
                )
                lifecycle = subprocess.run(
                    [str(args.server_tests.resolve()), "-test.run=^TestRealDatabaseLifecycle$", "-test.timeout=3m"],
                    cwd=root, env=environment, capture_output=True, text=True, timeout=190,
                )
                if lifecycle.returncode != 0:
                    raise AssertionError("database lifecycle tests failed:\n" + lifecycle.stdout + lifecycle.stderr)
                passed("Linux database lifecycle tests, including HTTPS and per-app mTLS")
            report["port"] = port
            report["passed"] = True
    finally:
        for service in reversed(services):
            service.stop()
        for name in reversed(containers):
            docker("rm", "-f", "-v", name)
    if args.report:
        args.report.parent.mkdir(parents=True, exist_ok=True)
        args.report.write_text(json.dumps(report, indent=2) + "\n")
    print(json.dumps(report, indent=2), flush=True)


if __name__ == "__main__":
    main()
