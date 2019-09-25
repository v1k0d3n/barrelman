package report

type Reportables interface {
	ShortReport() map[string]interface{}
	DetailedReport() map[string]interface{}
}
