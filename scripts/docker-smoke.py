"""Install and exercise only the dedicated auth-api-verification Compose project."""
import hashlib
import hmac
import json
import os
from pathlib import Path
import shutil
import subprocess
import time
import urllib.request
import uuid

ROOT = Path(__file__).resolve().parent.parent
BASE = "http://127.0.0.1:18080"
DOCKER = shutil.which("docker") or "C:/Program Files/Docker/Docker/resources/bin/docker.exe"
PROJECT = "auth-api-verification"
checks = []

def compose(*args):
    return subprocess.check_output([DOCKER, "compose", "-p", PROJECT, *args], cwd=ROOT, text=True)

def call(path, body=None, token="", app=None, sdk=False, expected=0):
    data = json.dumps(body, separators=(",", ":")).encode() if body is not None else None
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    if app:
        headers["X-App-Id"] = str(app["id"])
    if sdk:
        ts, nonce = str(int(time.time())), uuid.uuid4().hex
        signature = hmac.new(app["tokenKey"].encode(), ts.encode()+b"\n"+nonce.encode()+b"\n"+(data or b""), hashlib.sha256).hexdigest()
        headers.update({"X-Ts": ts, "X-Nonce": nonce, "X-Sign": signature, "X-Mi": "docker-device"})
    req = urllib.request.Request(BASE+path, data=data, headers=headers)
    with urllib.request.urlopen(req, timeout=90) as response:
        result = json.load(response)
    assert result["code"] == expected, (path, result)
    checks.append({"path": path, "code": result["code"]})
    return result.get("data")

def wait_ready():
    for _ in range(40):
        try:
            return call("/api/install/status")
        except (OSError, AssertionError):
            time.sleep(.5)
    raise RuntimeError("app did not become ready")

if __name__ == "__main__":
    (ROOT/"reports").mkdir(exist_ok=True)
    status = wait_ready()
    assert not status["installed"], "Smoke test requires a fresh disposable app volume"
    mysql = {"host":"mysql", "port":3306, "user":"root", "password":"local_test_only", "database":"auth_api", "tablePrefix":"a_"}
    redis = {"addr":"redis:6379", "db":0, "password":""}
    call("/api/install/test-mysql", mysql)
    call("/api/install/test-redis", redis)
    call("/api/install/save-config", {"mysql":mysql, "redis":redis})
    applied = call("/api/install/migrate", {})
    assert len(applied["applied"]) == 41, applied
    call("/api/install/create-admin", {"username":"admin", "password":"recovery-pass-123", "adminEntry":"private"})
    final = call("/api/install/finalize", {})
    assert final["restartRequired"] is False, final
    assert call("/api/install/status")["installed"]
    call("/api/install/migrate", {}, expected=403)
    admin = call("/api/private/login", {"username":"admin", "password":"recovery-pass-123"})["token"]
    app = call("/api/private/app", {"name":"Docker test", "type":"user", "mode":"free", "bindDevice":True}, token=admin)
    user = call("/api/sdk/register", {"username":"dockertest", "password":"test-password", "mi":"docker-device"}, app=app, sdk=True)
    call("/api/sdk/info", {}, token=user["token"], app=app, sdk=True)
    call("/api/private/dashboard", token=admin, app=app)
    assert compose("exec","-T","app","sh","-c","test ! -e data/master.key && echo absent").strip()=="absent"
    compose("restart","app")
    assert wait_ready()["installed"]
    call("/api/private/me", token=admin)
    call("/api/sdk/info", {}, token=user["token"], app=app, sdk=True)
    env = dict(os.environ, TEST_BASE_URL=BASE, TEST_APP_ID=str(app["id"]), TEST_REPORT_PATH=str(ROOT/"reports/docker-browser.json"))
    subprocess.run(["node","scripts/live-backend.mjs"], cwd=ROOT/"frontend", env=env, check=True)
    call("/api/sdk/logout", {}, token=user["token"], app=app, sdk=True)
    call("/api/sdk/info", {}, token=user["token"], app=app, sdk=True, expected=1006)
    (ROOT/"reports/docker-smoke.json").write_text(json.dumps({"passed":True, "platform":"linux-amd64", "install_migrations":41, "restart_session_persistence":True, "master_key_absent":True, "checks":checks}, indent=2), encoding="utf-8")
    print("Linux app install, SDK, restart persistence and browser smoke passed")
