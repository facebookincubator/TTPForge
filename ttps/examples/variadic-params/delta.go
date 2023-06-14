/*
Copyright © 2023-present, Meta Platforms, Inc. and affiliates
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/l50/goutils/v2/keeper"
	"github.com/l50/goutils/v2/str"
	"github.com/l50/goutils/v2/web"
	"github.com/l50/goutils/v2/web/cdpu"
	"github.com/pquerna/otp/totp"
	"github.com/spf13/cobra"
)

const (
	emailFieldXPath            = "/html/body/div/div/div/div/div/div/div/div/div[1]/div[2]/form/div/div/div/div/div/div/div/div[3]/div[2]/div[4]/div/div/div/div[2]/div/div/div[1]/div/div/input"
	passwordFieldXPath         = "/html/body/div/div/div/div/div/div/div/div/div[1]/div[2]/form/div/div/div/div/div/div/div/div[3]/div[2]/div[4]/div/div/div/div[2]/div/div/div[2]/div/div/input"
	loginButtonXPath           = "/html/body/div/div/div/div/div/div/div/div/div[1]/div[2]/form/div/div/div/div/div/div/div/div[4]/div[3]/div/div/div/div/div/div/div"
	loggedInXPath              = "/html/body/div/div[2]/div/div[2]/main/div/div[1]/div/div/div[1]/div/div[2]/div"
	pwConfirmFieldXPath        = "/html/body/div/div[2]/div/div/main/div/div/div[2]/div/div/div[1]/form/div/div[1]/div/div/div/input"
	pwConfirmLoginButtonXPath  = "/html/body/div/div[2]/div/div/main/div/div/div[2]/div/div/div[1]/form/div/div[2]/button"
	deleteAccountLinkXPath     = "/html/body/div/div[2]/div/div[2]/main/div/div[5]/div/div/a"
	confirmationCodeFieldXPath = `//input[@autocomplete='one-time-code']`
	nextButtonXPath            = `//div[@role='button' and .//span[contains(text(), 'Next')]]`
	allowAllCookiesButtonXPath = "/html/body/div[2]/div[1]/div/div[2]/div/div/div/div/div/div/div[3]/div[3]/div/div/div/div/div[1]/div[1]/div"
	okButtonXPath              = `//div[@role='button' and .//span[contains(text(), 'OK')]]`
)

var (
	deltaCmd = &cobra.Command{
		Use:   "delta",
		Short: "Execute login flow for account with delta 2FA triggered",
		Run: func(cmd *cobra.Command, args []string) {
			// Create TTP log
			logPath := fmt.Sprintf("%s.log", cmd.Use)
			if err := logging.InitLog(false, logPath, false, false); err != nil {
				logging.Logger.Sugar().Errorf("failed to initialize logger: %v", err)
				cobra.CheckErr(err)
			}

			logging.Logger.Sugar().Infof(
				"Executing %s - %s, please wait...",
				cmd.Use, cmd.Short)

			credential := web.Credential{
				User:       user,
				Password:   password,
				TwoFacCode: token,
			}

			var kr keeper.Record
			var err error
			// Get credentials from Keeper
			if keeperRecord != "" {
				kr, err = keeper.RetrieveRecord(keeperRecord)
				if err != nil {
					logging.Logger.Sugar().Error(err)
					cobra.CheckErr(err)
				}
				credential.User = kr.Username
				credential.Password = kr.Password
				if kr.TOTP != "" {
					// parse token
					parsedTOTPURL, err := url.Parse(kr.TOTP)
					if err != nil {
						logging.Logger.Sugar().Fatalf("failed to parse 2FA URL: %v", err)
						cobra.CheckErr(err)
					}
					queryParams, _ := url.ParseQuery(parsedTOTPURL.RawQuery)
					credential.TwoFacCode = queryParams.Get("secret")
				}
				logging.Logger.Sugar().Infof("Retrieved record with UID %s from keeper", keeperRecord)
			}

			// Initialize the chrome browser
			browser, err := cdpu.Init(headless, ignoreCertErrors)
			if err != nil {
				logging.Logger.Sugar().Errorf(
					"failed to initialize a chrome browser: %v", err)
				cobra.CheckErr(err)
			}
			defer web.CancelAll(browser.Cancels...)

			site := web.Site{
				LoginURL: target,
				Session: web.Session{
					Credential: credential,
					Driver:     browser.Driver,
				},
			}

			failure := false
			// Login to the target site
			if site.Session.Credential.TwoFacCode == "" {
				if err = LoginMetaAccount(site, web.WithTwoFac(false)); err != nil {
					logging.Logger.Sugar().Errorf("failed to login to %s: %v", site.LoginURL, err)
					failure = true
				}
			} else {
				logging.Logger.Sugar().Infof("Logging in with %s", site.Session.Credential.User)

				if err = LoginMetaAccount(site, web.WithTwoFac(true)); err != nil {
					logging.Logger.Sugar().Errorf("failed to login to %s: %v", site.LoginURL, err)
					failure = true
				}
			}

			// Necessary workaround for the browser to close. If we use cobra to check the
			// error, the browser will never close.
			if !failure {
				logging.Logger.Sugar().Infof("Successfully logged into %s with %s\n", site.LoginURL, site.Session.Credential.User)
			}
		},
	}

	keeperRecord     string
	ignoreCertErrors bool
	headless         bool
	password         string
	target           string
	user             string
	token            string
)

func init() {
	deltaCmd.Flags().StringVar(&target,
		"target", "https://auth.meta.com", "URL to delta to a meta account.")

	deltaCmd.Flags().BoolVar(&ignoreCertErrors,
		"ignoreCertErrors", false, "Ignore certificate errors.")

	deltaCmd.Flags().BoolVar(&headless,
		"headless", true, "Run browser in headless mode.")

	deltaCmd.Flags().StringVar(&user,
		"user", "", "Email address for the delta user")

	deltaCmd.Flags().StringVar(&password,
		"password", "", "Password for the delta user")

	deltaCmd.Flags().StringVar(&token,
		"token", "", "Delta user's 2FA token")

	deltaCmd.Flags().StringVar(&keeperRecord,
		"kr", "r0H-B6_g3PdEfVI4AAKMOw", "Path to the keeper file containing user credentials.")

}

func initialLogin(site web.Site) error {
	initialLoginActions := []cdpu.InputAction{
		{
			Description: "Navigate to the login page",
			Action:      chromedp.Navigate(site.LoginURL + "/login"),
		},
	}

	waitTime, err := web.GetRandomWait(6, 12)
	if err != nil {
		logging.Logger.Sugar().Errorf("failed to create random wait time: %v", err)
		return err
	}

	if err := cdpu.Navigate(site, initialLoginActions, waitTime); err != nil {
		logging.Logger.Sugar().Errorf("failed to navigate to %s: %v", site.LoginURL, err)
		return err
	}

	// Handle cookie prompt if it appears
	cookieMsg := "Allow the use of cookies from Meta on this browser?"
	attempts := 2
	for i := 0; i < attempts; i++ {
		pageSource, err := cdpu.GetPageSource(site)
		if err != nil {
			logging.Logger.Sugar().Error(err)
			return err
		}

		if strings.Contains(pageSource, cookieMsg) {
			initialLoginActions = []cdpu.InputAction{
				{
					Description: "Sleep to wait for the cookie prompt",
					Action:      chromedp.Sleep(5 * time.Second),
				},
				{
					Description: "Check if Allow all cookies button is present",
					Selector:    allowAllCookiesButtonXPath,
					Action:      chromedp.WaitVisible(allowAllCookiesButtonXPath),
				},
				{
					Description: "Click Allow all cookies",
					Selector:    allowAllCookiesButtonXPath,
					Action:      chromedp.Click(allowAllCookiesButtonXPath),
				},
			}

			err = cdpu.Navigate(site, initialLoginActions, waitTime)
			if err == nil {
				break
			}
			logging.Logger.Sugar().Errorf("failed to handle cookies (attempt %d): %v", i+1, err)
		}
		logging.Logger.Sugar().Debug("No cookie prompt")
	}

	return nil
}

func confirmLogin(site web.Site) error {
	confirmLoginAction := []cdpu.InputAction{
		{
			Description: "Wait for the DELETE YOUR ACCOUNT link to be present",
			Selector:    deleteAccountLinkXPath,
			Action:      chromedp.WaitVisible(deleteAccountLinkXPath),
		},
		{
			Description: "Sleep to give the site time to load",
			Action:      chromedp.Sleep(5 * time.Second),
		},
	}

	randomWaitTime, err := web.GetRandomWait(2, 6)
	if err != nil {
		logging.Logger.Sugar().Errorf("failed to create random wait time: %v", err)
		return err
	}

	if err := cdpu.Navigate(site, confirmLoginAction, randomWaitTime); err != nil {
		logging.Logger.Sugar().Errorf("failed to navigate to %s: %v", site.LoginURL, err)
		return err
	}

	imgPath := fmt.Sprintf("confirm-login-%s.png", site.Session.Credential.User)
	if err := cdpu.ScreenShot(site, imgPath); err != nil {
		logging.Logger.Sugar().Errorf("failed to navigate to %s: %v", site.LoginURL, err)
		return err
	}

	logging.Logger.Sugar().Infof("Successful login screenshot created for %s at %s", site.Session.Credential.User)

	return nil
}

func handlePostLogin(site web.Site) error {
	pageSource, err := cdpu.GetPageSource(site)
	if err != nil {
		return err
	}

	failMsg := "Login Failure"
	if strings.Contains(pageSource, failMsg) {
		err := fmt.Errorf("incorrect credential provided for %s", site.Session.Credential.User)
		return err
	}

	bruteForceMsg := "Confirm your Meta account"
	if strings.Contains(pageSource, bruteForceMsg) {
		return errors.New("2FA Brute force attack detected")
	}

	pageSource, err = cdpu.GetPageSource(site)
	if err != nil {
		return err
	}

	delta2FAMsg := "enter the 6-digit code we sent to"
	if strings.Contains(pageSource, delta2FAMsg) {
		err := errors.New("error: delta 2FA triggered")
		return err
	}

	return nil
}

func executeLoginActions(site web.Site) error {
	loginActions := []cdpu.InputAction{
		{
			Description: "Click Email field",
			Selector:    emailFieldXPath,
			Action:      chromedp.Click(emailFieldXPath),
		},
		{
			Description: "Input username",
			Selector:    emailFieldXPath,
			Action:      chromedp.SendKeys(emailFieldXPath, site.Session.Credential.User),
		},
		{
			Description: "Click Password field",
			Selector:    passwordFieldXPath,
			Action:      chromedp.Click(passwordFieldXPath),
		},
		{
			Description: "Input password",
			Selector:    passwordFieldXPath,
			Action:      chromedp.SendKeys(passwordFieldXPath, site.Session.Credential.Password),
		},
		{
			Description: "Click login button",
			Selector:    loginButtonXPath,
			Action:      chromedp.Click(loginButtonXPath),
		},
	}

	randomWaitTime, err := web.GetRandomWait(2, 6)
	if err != nil {
		logging.Logger.Sugar().Errorf("failed to create random wait time: %v", err)
		return err
	}

	if err := cdpu.Navigate(site, loginActions, randomWaitTime); err != nil {
		logging.Logger.Sugar().Errorf("failed to navigate to %s: %v", site.LoginURL, err)
		return err
	}

	return nil
}

func checkOkButton(site web.Site, done chan error) error {
	// Create a new context with a timeout
	chromeDriver, ok := site.Session.Driver.(*cdpu.Driver)
	if !ok {
		return errors.New("driver is not of type *Driver")
	}

	// Create a new context with a timeout
	ctx, cancel := context.WithTimeout(chromeDriver.GetContext(), 10*time.Second)
	defer cancel()

	actions := []cdpu.InputAction{
		{
			Description: "Check if the OK button is present - this signals the account is locked out",
			Selector:    okButtonXPath,
			Action: chromedp.ActionFunc(func(ctx context.Context) error {
				// Run the check for the OK button in a separate goroutine
				go func() {
					var nodes []*cdp.Node
					err := chromedp.Run(ctx, chromedp.Nodes(okButtonXPath, &nodes, chromedp.BySearch))
					if err == nil && len(nodes) > 0 {
						err = fmt.Errorf("%s account is locked out", site.Session.Credential.User)
					}
					// Send the result back to the main routine
					done <- err
				}()

				// Wait a while to allow the goroutine to complete
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Second * 5):
				}
				// Return nil here so chromedp doesn't see an error
				return nil
			}),
			Context: ctx,
		},
	}

	randomWaitTime, err := web.GetRandomWait(2, 6)
	if err != nil {
		logging.Logger.Sugar().Errorf("failed to create random wait time: %v", err)
		return err
	}

	return cdpu.Navigate(site, actions, randomWaitTime)
}

// LoginMetaAccount is a method that logs into a meta account.
//
// It can handle two-factor authentication and will log the user out after logging in
// if the corresponding options are set (by default, both are set to true).
//
// It takes in a web.Site object representing the site to be logged in to, as well
// as any number of LoginOption functions to modify the login options.
//
// Parameters:
//
// site: A web.Site struct containing the site to login to.
// options: Optional LoginOption functions to modify the login process.
//
// Returns:
//
// error: An error if one exists.
//
// By default, twoFacEnabled and logMeOut are set to true. To customize these options, use the WithTwoFac and WithLogout functions:
// - LoginMetaAccount(site) - both twoFacEnabled and logMeOut are set to their default values (true).
// - LoginMetaAccount(site, web.WithTwoFac(false)) - twoFacEnabled is set to false, logMeOut is set to its default value (true).
// - LoginMetaAccount(site, web.WithTwoFac(false)) - twoFacEnabled is set to its default value (true), logMeOut is set to false.
// - LoginMetaAccount(site, web.WithTwoFac(false), web.WithLogout(false)) - both twoFacEnabled and logMeOut are set to false.
func LoginMetaAccount(site web.Site, options ...web.LoginOption) error {
	loginOptions := web.SetLoginOptions(options...)

	if err := initialLogin(site); err != nil {
		return err
	}

	if err := executeLoginActions(site); err != nil {
		return err
	}

	// Create a channel to communicate the result of the OK button check
	done := make(chan error)
	if err := checkOkButton(site, done); err != nil {
		return err
	}

	logging.Logger.Sugar().Debugf("%s account is not locked out", site.Session.Credential.User)

	if web.IsTwoFacEnabled(loginOptions) {
		logging.Logger.Sugar().Infof("Two Factor enabled for the account - starting 2FA flow")
		if err := TwoFac(site); err != nil {
			logging.Logger.Sugar().Errorf("failed to get %s 2FA token: %v", site.Session.Credential.User, err)
			return err
		}
	}

	if err := handlePostLogin(site); err != nil {
		if strings.Contains(err.Error(), "2FA Brute force attack detected") {
			return nil
		}

		return err
	}

	if err := confirmLogin(site); err != nil {
		return err
	}

	if web.IsLogMeOutEnabled(loginOptions) {
		if err := LogoutMetaAccount(site); err != nil {
			logging.Logger.Sugar().Errorf("failed to log out of %s: %v", site.LoginURL, err)
			return err
		}
	}

	return nil
}

// LogoutMetaAccount logs out of a meta account.
//
// Parameters:
//
// site: A web.Site struct containing the site to logout from.
//
// Returns:
//
// error: An error if one exists.
func LogoutMetaAccount(site web.Site) error {
	logoutActions := []cdpu.InputAction{
		{
			Description: "Navigate to the logout endpoint",
			Action:      chromedp.Navigate(site.LoginURL + "/logout/"),
		},
	}

	waitTime, err := web.GetRandomWait(2, 6)
	if err != nil {
		logging.Logger.Sugar().Errorf("failed to create random wait time: %v", err)
		return err
	}

	if err := cdpu.Navigate(site, logoutActions, waitTime); err != nil {
		logging.Logger.Sugar().Errorf("failed to navigate to %s: %v", site.LoginURL, err)
		return err
	}

	logging.Logger.Sugar().Infof("Successfully logged out from %s", site.Session.Credential.User)

	return nil
}

// DeleteAccount facilitates deleting a meta account.
//
// Parameters:
//
// site: A web.Site struct containing the site to delete account from.
//
// Returns:
//
// error: An error if one exists.
func DeleteAccount(site web.Site) error {
	actions := []cdpu.InputAction{
		{
			Description: "Navigate to the delete account endpoint",
			Action:      chromedp.Navigate(site.LoginURL + "/settings/delete/"),
		},
		{
			Description: "Click the Password field",
			Selector:    pwConfirmFieldXPath,
			Action:      chromedp.Click(pwConfirmFieldXPath),
		},
		{
			Description: "Input the account password",
			Selector:    pwConfirmFieldXPath,
			Action:      chromedp.SendKeys(pwConfirmFieldXPath, site.Session.Credential.Password),
		},
		{
			Description: "Click the LOG IN button",
			Selector:    pwConfirmLoginButtonXPath,
			Action:      chromedp.Click(pwConfirmLoginButtonXPath),
		},
		{
			Description: "Click the DELETE YOUR ACCOUNT link",
			Selector:    deleteAccountLinkXPath,
			Action:      chromedp.Click(deleteAccountLinkXPath),
		},
	}

	waitTime, err := web.GetRandomWait(2, 6)
	if err != nil {
		logging.Logger.Sugar().Errorf("failed to create random wait time: %v", err)
		return err
	}

	if err := cdpu.Navigate(site, actions, waitTime); err != nil {
		logging.Logger.Sugar().Errorf("failed to execute actions: %v", err)
		return err
	}

	src, err := cdpu.GetPageSource(site)
	if err != nil {
		logging.Logger.Sugar().Errorf("failed to get page source: %v", err)
		return err
	}

	if strings.Contains(src, "Your account has been successfully scheduled for deletion") {
		logging.Logger.Sugar().Infof("Successfully scheduled %s for deletion!", site.Session.Credential.User)
	}

	return nil
}

// runInitialActions executes the logic to get to the 2FA input form.
//
// Parameters:
//
// site: A web.Site struct containing the site to execute actions on.
// randomWaitTime: A time.Duration struct containing the random wait time between actions.
//
// Returns:
//
// error: An error if one exists.
func startTwoFac(site web.Site, randomWaitTime time.Duration) error {
	twoFacMsg := "Please select a method below"
	pageSource, err := cdpu.GetPageSource(site)
	if err != nil {
		logging.Logger.Sugar().Error(err)
		return err
	}

	if strings.Contains(pageSource, twoFacMsg) {
		// Prepare initial actions
		initialActions := []cdpu.InputAction{
			{
				Description: "Click the Next button",
				Selector:    nextButtonXPath,
				Action:      chromedp.Click(nextButtonXPath),
			},
		}

		// Run initial actions to get to the 2FA input form
		if err := cdpu.Navigate(site, initialActions, randomWaitTime); err != nil {
			logging.Logger.Sugar().Error(err)
			return err
		}

	}

	return nil
}

// SubmitTwoFacCode submits a 2FA token and determines if the token
// provided a successful login.
//
// Parameters:
//
// site: A web.Site struct containing the site to submit the 2FA token to.
// token: A string representing the 2FA token.
// randomWaitTime: A time.Duration struct containing the random wait time between actions.
//
// Returns:
//
// error: An error if one exists.
func SubmitTwoFacCode(site web.Site, randomWaitTime time.Duration) error {
	if err := submitCode(site, randomWaitTime); err != nil {
		return err
	}

	attempts := 2

	for i := 0; i < attempts; i++ {
		pageSource, err := cdpu.GetPageSource(site)
		if err != nil {
			logging.Logger.Sugar().Error(err)
			return err
		}

		// Handle incorrect 2FA code
		invalidTokenMsg := "Incorrect Login Code"
		if strings.Contains(pageSource, invalidTokenMsg) {
			if err := fmt.Errorf("invalid 2FA code %s", site.Session.Credential.TwoFacCode); err != nil {
				logging.Logger.Sugar().Error(err)
				return err
			}

			// Check for OK button existence
			var nodes []*cdp.Node

			actions := []cdpu.InputAction{
				{
					Description: "Check if the OK button is present - this signals the code failed",
					Selector:    okButtonXPath,
					Action:      chromedp.Nodes(okButtonXPath, &nodes, chromedp.BySearch),
				},
			}

			if err := cdpu.Navigate(site, actions, randomWaitTime); err != nil {
				logging.Logger.Sugar().Error(err)
				return err
			}

			if len(nodes) > 0 {
				err := fmt.Errorf("incorrect 2FA code %s provided by %s", site.Session.Credential.TwoFacCode, site.Session.Credential.User)
				logging.Logger.Sugar().Error(err)
				return err
			}
		}
	}

	return nil
}

func checkSMSLockout(site web.Site) error {
	lockoutCheckActions := []cdpu.InputAction{
		{
			Description: "Sleep to give the site time to load",
			Action:      chromedp.Sleep(5 * time.Second),
		},
	}

	if err := cdpu.Navigate(site, lockoutCheckActions, 5); err != nil {
		return err
	}

	pageSource, err := cdpu.GetPageSource(site)
	if err != nil {
		logging.Logger.Sugar().Error(err)
		return err
	}
	smsLockoutMsg := "You've requested too many codes"
	if strings.Contains(pageSource, smsLockoutMsg) {
		err = fmt.Errorf("can't request additional SMS 2FA tokens - " +
			"wait for a bit and try again later")
		logging.Logger.Sugar().Error(err)
		return err
	}

	return nil
}

// submitCode submits a 2FA token for the associated user.
//
// Parameters:
//
// site: A web.Site struct containing the site to submit the token to.
// token: A string representing the token.
// randomWaitTime: A time.Duration struct containing the random wait time between actions.
//
// Returns:
//
// error: An error if one exists.
func submitCode(site web.Site, randomWaitTime time.Duration) error {
	if err := checkSMSLockout(site); err != nil {
		return err
	}

	token := site.Session.Credential.TwoFacCode
	if len(token) != 6 || !str.IsNumeric(token) {
		var err error
		token, err = totp.GenerateCode(site.Session.Credential.TwoFacCode, time.Now())
		if err != nil {
			logging.Logger.Sugar().Errorf("failed to generate 2FA token: %v", err)
			return err
		}
	}

	pageSource, err := cdpu.GetPageSource(site)
	if err != nil {
		logging.Logger.Sugar().Error(err)
		return err
	}
	twoFacMsg := "Enter the code"
	if strings.Contains(pageSource, twoFacMsg) {
		actions := []cdpu.InputAction{
			{
				Description: "Wait for the Confirmation code input field to become available",
				Selector:    confirmationCodeFieldXPath,
				Action:      chromedp.WaitVisible(confirmationCodeFieldXPath, chromedp.BySearch),
			},
			{
				Description: "Click the Confirmation code input field",
				Selector:    confirmationCodeFieldXPath,
				Action:      chromedp.Click(confirmationCodeFieldXPath),
			},
			{
				Description: "Clear the Confirmation code input field",
				Selector:    confirmationCodeFieldXPath,
				Action:      chromedp.SendKeys(confirmationCodeFieldXPath, "\b\b\b\b\b\b"),
			},
			{
				Description: "Input the 2FA code",
				Selector:    confirmationCodeFieldXPath,
				Action:      chromedp.SendKeys(confirmationCodeFieldXPath, token),
			},
			{
				Description: "Wait for the Next button to be ready",
				Selector:    nextButtonXPath,
				Action:      chromedp.WaitReady(nextButtonXPath, chromedp.BySearch),
			},
			{
				Description: "Click the Next button to submit the input confirmation code",
				Selector:    nextButtonXPath,
				Action:      chromedp.Click(nextButtonXPath),
			},
			{
				Description: "Sleep to give site time to load",
				Action:      chromedp.Sleep(5 * time.Second),
			},
		}

		// Run actions to submit 2FA code
		if err := cdpu.Navigate(site, actions, randomWaitTime); err != nil {
			return err
		}
	}

	return nil
}

// TwoFac executes two-factor logic for the input user.
//
// Parameters:
//
// site: A web.Site struct containing the site to execute the two-factor logic on.
//
// Returns:
//
// error: An error if one exists.
func TwoFac(site web.Site) error {
	waitTime, err := web.GetRandomWait(2, 6)
	if err != nil {
		logging.Logger.Sugar().Errorf("failed to create random wait time: %v", err)
		return err
	}

	if err := startTwoFac(site, waitTime); err != nil {
		return err
	}

	if err := SubmitTwoFacCode(site, waitTime); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := deltaCmd.Execute(); err != nil {
		logging.Logger.Sugar().Errorf("%s failed to run: %v", deltaCmd.Short, err)
		os.Exit(1)
	}
}
