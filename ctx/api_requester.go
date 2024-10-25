package ctx

import "net/http"

func SetHTTPRequestHeader(r *http.Request) {
	mt := Get(r.Context())
	r.Header.Set("X-Request-ID", mt.ReqId)
	r.Header.Set("X-Request-User", mt.ReqUser)
	r.Header.Set("X-Api-Key", mt.ApiKey)
	r.Header.Set("Authorization", mt.Authorization)
	r.Header.Set("X-Path-Gateway", mt.PathGateway)
	r.Header.Set("X-Request-Kode-Region", mt.ReqKodeRegion)
	r.Header.Set("X-Request-Kode-MainUker", mt.ReqKodeMainUker)
	r.Header.Set("X-Request-Jenis-Uker", mt.ReqJenisUker)
	r.Header.Set("X-Request-Nama-Uker", mt.ReqNamaUker)
	r.Header.Set("X-Request-Kode-Uker", mt.ReqKodeUker)
	r.Header.Set("X-Request-Nama-Jabatan", mt.ReqNamaJabatan)
	r.Header.Set("X-Request-Kode-Jabatan", mt.ReqKodeJabatan)
	r.Header.Set("X-Request-Nama", mt.ReqNama)
	r.Header.Set("X-Request-Tags", mt.ReqTags)
	r.Header.Set("X-Version-App", mt.VersionApp)
	r.Header.Set("X-App", mt.App)
	r.Header.Set("X-Device-ID", mt.DeviceId)
	r.Header.Set("X-Request-From", mt.RequestFrom)
	r.Header.Set("X-Forwarded-For", mt.ForwardedFor)
	r.Header.Set("Cache-Control", mt.CacheControl)
	r.Header.Set("User-Agent", mt.UserAgent)
}
