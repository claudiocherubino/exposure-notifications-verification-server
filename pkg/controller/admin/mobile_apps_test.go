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

package admin_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/exposure-notifications-verification-server/internal/browser"
	"github.com/google/exposure-notifications-verification-server/internal/envstest"
	"github.com/google/exposure-notifications-verification-server/pkg/controller"
	"github.com/google/exposure-notifications-verification-server/pkg/database"

	"github.com/chromedp/chromedp"
)

func TestShowAdminMobileApps(t *testing.T) {
	t.Parallel()

	harness := envstest.NewServer(t, testDatabaseInstance)

	// Get the default realm
	realm, err := harness.Database.FindRealm(1)
	if err != nil {
		t.Fatal(err)
	}

	// Create a system admin
	admin := &database.User{
		Email:       "admin@example.com",
		Name:        "Admin",
		SystemAdmin: true,
		Realms:      []*database.Realm{realm},
		AdminRealms: []*database.Realm{realm},
	}
	if err := harness.Database.SaveUser(admin, database.SystemTest); err != nil {
		t.Fatal(err)
	}

	app := &database.MobileApp{
		Name:    "test mobile app",
		RealmID: realm.ID,
		URL:     "https://example2.com",
		OS:      database.OSTypeAndroid,
		SHA:     "AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA:AA",
		AppID:   "app2",
	}
	if err := harness.Database.SaveMobileApp(app, database.SystemTest); err != nil {
		t.Fatal(err)
	}

	// Log in the user.
	session, err := harness.LoggedInSession(nil, admin.Email)
	if err != nil {
		t.Fatal(err)
	}

	// Set the current realm.
	controller.StoreSessionRealm(session, realm)

	// Mint a cookie for the session.
	cookie, err := harness.SessionCookie(session)
	if err != nil {
		t.Fatal(err)
	}
	// Create a browser runner.
	browserCtx := browser.New(t)
	taskCtx, done := context.WithTimeout(browserCtx, 30*time.Second)
	defer done()

	if err := chromedp.Run(taskCtx,
		// Pre-authenticate the user.
		browser.SetCookie(cookie),

		// Visit /admin
		chromedp.Navigate(`http://`+harness.Server.Addr()+`/admin/mobile-apps`),

		// Wait for render.
		chromedp.WaitVisible(`body#admin-mobileapps-index`, chromedp.ByQuery),

		// Fill out the form by email.
		chromedp.SetValue(`input#search`, "test mobile app", chromedp.ByQuery),
		chromedp.Submit(`form#search-form`, chromedp.ByQuery),

		// Wait for the search result.
		chromedp.WaitVisible(`table#results-table tr`, chromedp.ByQuery),

		// Fill out the form with a non-existing user
		chromedp.SetValue(`input#search`, "notexists", chromedp.ByQuery),
		chromedp.Submit(`form#search-form`, chromedp.ByQuery),

		// Assert no users shown
		chromedp.WaitNotPresent(`table#results-table tr`, chromedp.ByQuery),
	); err != nil {
		t.Fatal(err)
	}
}
