package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

const (
	comma = ','
	equal = '='
)

var keyValueRegex = regexp.MustCompile(`([\w-:;./@]+=([\w-:;,./@][\w-:;,./@ ]*[\w-:;,./@])?[\w-:;,./@]*,)*([\w-:;./@]+=([\w-:;,./@][\w-:;,./@ ]*)?[\w-:;./@]+)$`)

func GetStringMapStringE(vp *viper.Viper, key string) (map[string]string, error) {
	return ToStringMapStringE(vp.Get(key))
}

func ToStringMapStringE(data interface{}) (map[string]string, error) {
	if data == nil {
		return map[string]string{}, nil
	}

	v, err := cast.ToStringMapStringE(data)
	if err != nil {
		var jsonsyntaxErr *json.SyntaxError
		if !errors.As(err, &jsonsyntaxErr) {
			return v, err
		}

		switch s := data.(type) {
		case string:
			if len(s) == 0 {
				return map[string]string{}, nil
			}

			firstIndex := strings.IndexFunc(s, func(r rune) bool {
				return !unicode.IsSpace(r)
			})
			if firstIndex != -1 && (s[firstIndex] == '{' || s[firstIndex] == '[') {
				return v, err
			}
			if !isValidKeyValuePair(s) {
				return map[string]string{}, fmt.Errorf("invalid string format")
			}
			var v = map[string]string{}
			kvs := splitKeyValue(s, comma, equal)
			for _, kv := range kvs {
				tmp := strings.Split(kv, string(equal))
				if len(tmp) == 0 {
					return map[string]string{}, fmt.Errorf("invalid map string format")
				}
				v[tmp[0]] = tmp[1]
			}
			return v, nil
		}

	}
	return v, nil
}

// hard generic function
func isValidKeyValuePair(str string) bool {
	if len(str) == 0 {
		return true
	}
	return len(keyValueRegex.ReplaceAllString(str, "")) == 0
}

func splitKeyValue(str string, sep rune, keyValueSep rune) []string {
	var sepIndexes, kvValueSepIndexes []int
	// find all indexes of separator character
	for i := 0; i < len(str); i++ {
		switch int32(str[i]) {
		case sep:
			sepIndexes = append(sepIndexes, i)
		case keyValueSep:
			kvValueSepIndexes = append(kvValueSepIndexes, i)
		}
	}

	if len(sepIndexes) == 0 || len(kvValueSepIndexes) == 1 {
		return []string{str}
	}

	if len(sepIndexes) == 1 {
		index := sepIndexes[0]
		return []string{str[:index], str[index+1:]}
	}

	var res []string
	var start = 0
	for i := 0; i < len(sepIndexes); i++ {
		last := len(str)
		if i < len(sepIndexes)-1 {
			last = sepIndexes[i+1]
		}
		if strings.ContainsRune(str[sepIndexes[i]:last], keyValueSep) {
			res = append(res, str[start:sepIndexes[i]])
			start = sepIndexes[i] + 1
		}
	}
	res = append(res, str[start:])
	return res
}
