//golangcitest:args -Emusttag
//golangcitest:config_path testdata/configs/musttag.yml
package testdata

import (
	"encoding/asn1"
	"encoding/json"
)

// builtin functions:
func musttagJSON() {
	var user struct { // want `\Qexported fields should be annotated with the "json" tag`
		Name  string
		Email string `json:"email"`
	}
	json.Marshal(user)
	json.Unmarshal(nil, &user)
}

// custom functions from config:
func musttagCustom() {
	var user struct { // want `\Qexported fields should be annotated with the "asn1" tag`
		Name  string
		Email string `asn1:"email"`
	}
	asn1.Marshal(user)
	asn1.Unmarshal(nil, &user)
}
