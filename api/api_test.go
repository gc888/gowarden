package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"strings"

	"github.com/404cn/gowarden/sqlite/mock"
)

var mux *http.ServeMux
var writer *httptest.ResponseRecorder
var testHandler = New(mock.New())

func TestMain(m *testing.M) {
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func setUp() {
	mux = http.NewServeMux()
	mux.HandleFunc("/api/accounts/register", testHandler.HandleRegister)
	writer = httptest.NewRecorder()
}

func tearDown() {}

func TestHandleRegister(t *testing.T) {
	json := strings.NewReader(`{
"name": "",
"email": "nobody@example.com",
"masterPasswordHash": "LYYuwVfz5q+wj/9JYqlfidzOf/ytIOC2X3EDTspXmvk=",
"masterPasswordHint": "example",
"key": "2.OdNZrhgvjJLvTtC3vf8qAA==|18q18HPSm4J87x5zUjISnnGm9EfuA3Ljb6G68FdpmSHWGLS3IzZQY5McrckMSj36qJY+QLQdBVrABldFn6O8wfxU3j70JI/o4ySqvK07jBk=|+lNuLsPz3Yg9MsTlQfkSIBbAy4EJVsPdx3gf991OA5E=",
"kdf": 0,
"kdfIterations": 100000,
"keys": {
"publicKey": "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAs3qcoieRPRdep2VXIthP0ensfEfZyTKehbGJm0BggOgl7AXwWDNf561eEkBolDclObmKBinL9O0WD2qk5jVNK53tdWEsXeZhT6jNFXpNqpucYsHguUyfzhFPU8PtITsOr4PrAb/qG5g2z+6APKCsVXTzbSyw5jXgj6TuPfuMgt7LFe4qvbdU4FCQ1Fi23ZSfywxOpvMb7oarkc/eEYWCgX63p+mauqk3I1AEeJAwdozIJpXv7S0qpNQ2RgTi2LlQs42rR87dp6QT4z7BaDAXinzOf2lfs8hV6e4O3S6NURhzfprR4ovWLzzo3Bd2VJYwDha+93ljtdXN4Uo78+BvwwIDAQAB",
"encryptedPrivateKey": "2.rnMY8o1Cjh2pKSzZwGG0+w==|l0VfO+SxuTxSU4PMwX/8srbb3g0izV61O91YfqBU/FNI/bnECLvO5mM/nytdJ5e/vnauEnashvYhp4fr86oDfeWXSSO2V54qNqu1xS0mWbX1S0/uhy+uOSLL7U7m7vqG2aL3r8fl/VGjTdavfRwOMlya09hnRaeekVUePUQiMlpMayd5iSQ/T3unpI8Nn6I41WtEzyeVOvMQ05+Upd+7pBxfa16M2OFKC9olNgOEhE2X8qT6xx6w/jZcb/VUHdhTfilbI5nDtGDCZLJCQbS83cV2U3sHEEmarafEBFEx1rjscph7YuDT5CCy5vYB7OG4TbodQVwW3q/S6g8q05COc1DEeROspzv1BiSiB0rSuMdkoFbMBLYGrqn4xWQtjOoZF7b07o3RLGtg7JRUq+nNyyv8pnwa9yrVnwz1nYNZPhzDE/Bo3lFu95M+W3YXnV9ny2YbNX+dpk38hdiRjQqh8sT8Uw0g9GbOdXtAxL71SCzfqXGEusyUVXYI7WUju5nqljI41BQq2niks6t8x6xTPPD+NM0TjcFAMYcKaezW5MEEEKsNIE0kZYmrjnRp+4JFVEfI1VQdM518yWkEC9coaESOwai0Spz0OsrsfEu1oLgWkdkd/o6uwf4yLyuc+nwSpqhh96oS34FCS0r97xPMSYf49kR+4Tge/kDbILlbmAIzis2Qqqy+Q9OJSNP/aOqqqCBAyXn76Y9mvc0+Ui2BZONegIA/kvoH5sDCZNRnrjOhlkkGEHQ4jYHwXo6kNI6XzIejG2tZnuKI9ocKlaa+YZzAEq3WGoKN3PwIfCsWLg5dxa1M0HUa6xCqylhy79d/yGmR2vAF+RfusU4z1FeV2X+6QZVjTPYeNPCwJZx7dLDH1zFnuhHFtBeoiuxR78xM6xLjFjL/6EiSO/qAmz4WvShGLBQl5O/pcTJK4Fe9GVO2XH/zolsgNdjN7Co/99YIju4uiS5g/A+DpJgBkBorJYuuF4UPmcr62NobzVnL5UlZgPht1EWna2usMrOzCwkHocI5r/E8/ng1oXxD31WXSg5IcAuhEIdUQPsZemjPQYZa0BSWjO3nXv/qjmSk8wM4GHUEjebvV3X2U4mYr278kwCD+323aoE97u+1eWzbgTVzp0vDR3FsbYD76bMz9g/s4ZVnGZck7wUxpx0mlUpr9ja//o7vo7wrRyT/UQcUX57yy2B7wbZzwPDmBLx6x4bcEzlaZpYDr3f4aGiweNYO0f1FqiwBJ6GyBg9FMhKkdRyghq82bALWWd2P+OeH2A0Zd2J5HJOEEJ13wCFhaodQcDQRu47paCtpRtZws5qHjCUY3TIioDHmV0Rj+WioRAE41MXqeFX29xH7xE8VZaxnZldqNbIRaYVXnMAw24Z3uCxk8pAlegBrgTaOmn8xmZEPFeOe2QURaBuOhh2nFzjmGxqjvfyVSLZqck7Syx3cPl/S61VsROVK2+E8gl8MOaFa8H4CPPa9pSDiHG1Ua0Kp3tb5/2wFj/aQuMvFG1MGsSSYZ2li1nCVLJRwWwsCpwEVf9kdz08+zArZF9CLSGjkS8OKF6SbRmp8FuKXaU/FgjjK3+Tj9kb1g2pb4d6a4EH1hwau0iRfKHyFUzlRcE3RQgB466d2EbY/9SE8nCaanjA=|VB2SILyh9/uo02DtsvhFjSYkog/O8NirEaGljw42YjU="
}
}`)
	// TODO nil -> test json
	request, _ := http.NewRequest("POST", "/api/accounts/register", json)
	mux.ServeHTTP(writer, request)

	if writer.Code != 200 {
		t.Errorf("Response code is %v", writer.Code)
	}
}
