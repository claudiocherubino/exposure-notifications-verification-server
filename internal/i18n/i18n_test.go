// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package i18n

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/exposure-notifications-verification-server/internal/project"
	"github.com/leonelquinteros/gotext"
)

// TestI18n_matching constructs the superset of all i18n strings and then
// ensures all translation files define said strings.
func TestI18n_matching(t *testing.T) {
	t.Parallel()

	var pos []*gotext.Po
	localesDir := filepath.Join(project.Root(), "internal", "i18n", "locales")
	if err := filepath.Walk(localesDir, func(pth string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(info.Name()) != ".po" {
			return nil
		}

		po := new(gotext.Po)
		po.ParseFile(pth)
		pos = append(pos, po)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	// This will almost certainly come back to bite me, but the only way to access
	// the actual "list" of translations is to access a private field with
	// reflect. Please don't try this at home kids.
	translations := make(map[string]struct{})
	translationsByLocale := make(map[string]map[string]struct{})
	for _, po := range pos {
		keys := reflect.ValueOf(po).Elem().FieldByName("translations").MapKeys()
		for _, v := range keys {
			if s := v.String(); s != "" {
				translations[s] = struct{}{}

				if translationsByLocale[po.Language] == nil {
					translationsByLocale[po.Language] = make(map[string]struct{})
				}
				translationsByLocale[po.Language][s] = struct{}{}
			}
		}
	}

	for k := range translations {
		for loc, existing := range translationsByLocale {
			if _, ok := existing[k]; !ok {
				t.Errorf("locale %q is missing translation %q", loc, k)
			}
		}
	}
}
