// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package audit

import (
	"regexp"
	"strings"

	"github.com/nbutton23/zxcvbn-go"

	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func isUpperRune(r rune) bool {
	// NOTE: Intentionally not use unicode.IsUpper
	// because it take other languages into account.
	return r >= 'A' && r <= 'Z'
}

func isLowerRune(r rune) bool {
	// NOTE: Intentionally not use unicode.IsLower
	// because it take other languages into account.
	return r >= 'a' && r <= 'z'
}

func isDigitRune(r rune) bool {
	// NOTE: Intentionally not use unicode.IsDigit
	// because it take other languages into account.
	return r >= '0' && r <= '9'
}

func isSymbolRune(r rune) bool {
	// We define symbol as non-alphanumeric character
	return !isUpperRune(r) && !isLowerRune(r) && !isDigitRune(r)
}

func checkPasswordLength(password string, minLength int) bool {
	if minLength <= 0 {
		return true
	}
	// There exist many ways to define the length of a string
	// For example:
	// 1. The number of bytes of a given encoding
	// 2. The number of code points
	// 3. The number of extended grapheme cluster
	// Here we use the simpliest one:
	// the number of bytes of the given string in UTF-8 encoding
	return len(password) >= minLength
}

func checkPasswordUppercase(password string) bool {
	for _, r := range password {
		if isUpperRune(r) {
			return true
		}
	}
	return false
}

func checkPasswordLowercase(password string) bool {
	for _, r := range password {
		if isLowerRune(r) {
			return true
		}
	}
	return false
}

func checkPasswordDigit(password string) bool {
	for _, r := range password {
		if isDigitRune(r) {
			return true
		}
	}
	return false
}

func checkPasswordSymbol(password string) bool {
	for _, r := range password {
		if isSymbolRune(r) {
			return true
		}
	}
	return false
}

func checkPasswordExcludedKeywords(password string, keywords []string) bool {
	if len(keywords) <= 0 {
		return true
	}
	words := []string{}
	for _, w := range keywords {
		words = append(words, regexp.QuoteMeta(w))
	}
	re, err := regexp.Compile("(?i)" + strings.Join(words, "|"))
	if err != nil {
		return false
	}
	loc := re.FindStringIndex(password)
	if loc == nil {
		return true
	}
	return false
}

func checkPasswordGuessableLevel(password string, minLevel int, userInputs []string) (int, bool) {
	if minLevel <= 0 {
		return 0, true
	}
	minScore := minLevel - 1
	if minScore > 4 {
		minScore = 4
	}
	result := zxcvbn.PasswordStrength(password, userInputs)
	ok := result.Score >= minScore
	return result.Score + 1, ok
}

func userDataToStringStringMap(m map[string]interface{}) map[string]string {
	output := make(map[string]string)
	for key, value := range m {
		str, ok := value.(string)
		if ok {
			output[key] = str
		}
	}
	return output
}

func filterDictionary(m map[string]string, predicate func(string) bool) []string {
	output := []string{}
	for key, value := range m {
		ok := predicate(key)
		if ok {
			output = append(output, value)
		}
	}
	return output
}

func filterDictionaryByKeys(m map[string]string, keys []string) []string {
	lookupMap := make(map[string]bool)
	for _, key := range keys {
		lookupMap[key] = true
	}
	predicate := func(key string) bool {
		_, ok := lookupMap[key]
		return ok
	}

	return filterDictionary(m, predicate)
}

func filterDictionaryTakeAll(m map[string]string) []string {
	predicate := func(key string) bool {
		return true
	}
	return filterDictionary(m, predicate)
}

type UserAuditor struct {
	Enabled             bool
	PwMinLength         int
	PwUppercaseRequired bool
	PwLowercaseRequired bool
	PwDigitRequired     bool
	PwSymbolRequired    bool
	PwMinGuessableLevel int
	PwExcludedKeywords  []string
	PwExcludedFields    []string
	PwHistorySize       int
	PwHistoryDays       int
	PwExpiryDays        int
}

func (ua *UserAuditor) checkPasswordLength(password string) skyerr.Error {
	minLength := ua.PwMinLength
	if minLength > 0 && !checkPasswordLength(password, minLength) {
		return skyerr.NewErrorWithInfo(
			skyerr.PasswordTooShort,
			"password too short",
			map[string]interface{}{
				"min_length": minLength,
				"pw_length":  len(password),
			},
		)
	}
	return nil
}

func (ua *UserAuditor) checkPasswordUppercase(password string) skyerr.Error {
	if ua.PwUppercaseRequired && !checkPasswordUppercase(password) {
		return skyerr.NewError(
			skyerr.PasswordUppercaseRequired,
			"password uppercase required",
		)
	}
	return nil
}

func (ua *UserAuditor) checkPasswordLowercase(password string) skyerr.Error {
	if ua.PwLowercaseRequired && !checkPasswordLowercase(password) {
		return skyerr.NewError(
			skyerr.PasswordLowercaseRequired,
			"password lowercase required",
		)
	}
	return nil
}

func (ua *UserAuditor) checkPasswordDigit(password string) skyerr.Error {
	if ua.PwDigitRequired && !checkPasswordDigit(password) {
		return skyerr.NewError(
			skyerr.PasswordDigitRequired,
			"password digit required",
		)
	}
	return nil
}

func (ua *UserAuditor) checkPasswordSymbol(password string) skyerr.Error {
	if ua.PwSymbolRequired && !checkPasswordSymbol(password) {
		return skyerr.NewError(
			skyerr.PasswordSymbolRequired,
			"password symbol required",
		)
	}
	return nil
}

func (ua *UserAuditor) checkPasswordExcludedKeywords(password string) skyerr.Error {
	keywords := ua.PwExcludedKeywords
	if len(keywords) > 0 && !checkPasswordExcludedKeywords(password, keywords) {
		return skyerr.NewError(
			skyerr.PasswordContainingExcludedKeywords,
			"password containing excluded keywords",
		)
	}
	return nil
}

func (ua *UserAuditor) checkPasswordExcludedFields(password string, userData map[string]interface{}) skyerr.Error {
	fields := ua.PwExcludedFields
	if len(fields) > 0 {
		dict := userDataToStringStringMap(userData)
		keywords := filterDictionaryByKeys(dict, fields)
		if !checkPasswordExcludedKeywords(password, keywords) {
			return skyerr.NewError(
				skyerr.PasswordContainingExcludedKeywords,
				"password containing excluded keywords",
			)
		}
	}
	return nil
}

func (ua *UserAuditor) checkPasswordGuessableLevel(password string, userData map[string]interface{}) skyerr.Error {
	minLevel := ua.PwMinGuessableLevel
	if minLevel > 0 {
		dict := userDataToStringStringMap(userData)
		userInputs := filterDictionaryTakeAll(dict)
		level, ok := checkPasswordGuessableLevel(password, minLevel, userInputs)
		if !ok {
			return skyerr.NewErrorWithInfo(
				skyerr.PasswordBelowGuessableLevel,
				"password below guessable level",
				map[string]interface{}{
					"min_level": minLevel,
					"pw_level":  level,
				},
			)
		}
	}
	return nil
}

func (ua *UserAuditor) ValidatePassword(password string, userData map[string]interface{}) skyerr.Error {
	if err := ua.checkPasswordLength(password); err != nil {
		return err
	}
	if err := ua.checkPasswordUppercase(password); err != nil {
		return err
	}
	if err := ua.checkPasswordLowercase(password); err != nil {
		return err
	}
	if err := ua.checkPasswordDigit(password); err != nil {
		return err
	}
	if err := ua.checkPasswordSymbol(password); err != nil {
		return err
	}
	if err := ua.checkPasswordExcludedKeywords(password); err != nil {
		return err
	}
	if err := ua.checkPasswordExcludedFields(password, userData); err != nil {
		return err
	}
	return ua.checkPasswordGuessableLevel(password, userData)
}

func (ua *UserAuditor) ShouldSavePasswordHistory() bool {
	return ua.PwHistorySize > 0 || ua.PwHistoryDays > 0
}
