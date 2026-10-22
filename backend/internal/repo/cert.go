package repo

import "context"

func (r *AppRepo) CACertificate(ctx context.Context, id int64) (string, error) {
	var v string
	e := r.db.GetContext(ctx, &v, SQL("SELECT COALESCE(mtls_ca_cert,'') FROM {apps} WHERE id=?"), id)
	return v, e
}
