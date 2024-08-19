package stderr

// The standard error codes that is used in brispot.
const (
	ERROR_CODE_INVALID_HEADER    = "X0"
	ERROR_CODE_ACCESS_PERMISSION = "X1"
	ERROR_CODE_PARAMETER         = "01" // 200, mostly in controller when failed in validation rules
	ERROR_CODE_DATA_NOT_FOUND    = "02"
	ERROR_CODE_INVALID_RULE      = "03"
	ERROR_CODE_THIRD_PARTY       = "04"
	ERROR_CODE_WAITING_STATUS    = "05"
	ERROR_CODE_UNSUPPORTED       = "06"
	ERROR_CODE_SYSTEM            = "99" // 200, e.g. failed in connection
)

// The standard error description that is used in brispot.
const (
	ERROR_DESC_INVALID_HEADER = "Header request tidak valid"
	ERROR_DESC_PERMISSION     = "Akses tidak diijinkan"
	ERROR_DESC_PARAMETER      = "Parameter tidak sesuai"
	ERROR_DESC_DATA_NOT_FOUND = "Data tidak ditemukan"
	ERROR_DESC_INVALID_RULE   = "Validasi rule tidak terpenuhi"
	ERROR_DESC_THIRD_PARTY    = "Service ThirdParty bermasalah"
	ERROR_DESC_WAITING_STATUS = "Masih proses harap tunggu"
	ERROR_DESC_UNSUPPORTED    = "Mekanisme tidak disupport"
	ERROR_DESC_SYSTEM         = "Runtime error happens"
)
