import {Link, useNavigate} from "react-router-dom";
import {getRandomName} from "constants/name";
import {Auth} from "utils/auth";
import {Toast} from "utils/Toast";
import {useState} from "react";
import {LoginProviders} from "components/LoginProviders";
import {Trans, useTranslation} from "react-i18next";
import {useLocation} from "react-router";
import {HeroIllustration} from "components/HeroIllustration";
import {ScrumlrLogo} from "components/ScrumlrLogo";
import {ReactComponent as RefreshIcon} from "assets/icon-refresh.svg";
import {ReactComponent as KeyIcon} from "assets/icon-key.svg";
import "./LoginBoard.scss";
import {TextInputAction} from "components/TextInputAction";
import {Button} from "components/Button";
import {TextInput} from "components/TextInput";
import {TextInputLabel} from "components/TextInputLabel";
import {ValidationError} from "components/ValidationError";
import {startAuthentication, startRegistration} from "@simplewebauthn/browser";
import {SERVER_HTTP_URL, SHOW_LEGAL_DOCUMENTS} from "../../config";

interface State {
  from: {pathname: string};
}

export const LoginBoard = () => {
  const {t} = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();

  // const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState(getRandomName());
  const [termsAccepted, setTermsAccepted] = useState(!SHOW_LEGAL_DOCUMENTS);
  const [submitted, setSubmitted] = useState(false);

  let redirectPath = "/";
  if (location.state) {
    redirectPath = (location.state as State).from.pathname;
  }

  // useEffect(() => {
  //   loginWithPasskey(true) //rename to startLogin
  // }, []);

  // anonymous sign in and redirection to board path that is in history
  async function handleLogin() {
    if (termsAccepted) {
      try {
        await Auth.signInAnonymously(displayName);
        navigate(redirectPath);
      } catch (err) {
        Toast.error({title: t("LoginBoard.errorOnRedirect")});
      }
    }
    setSubmitted(true);
  }

  async function registerPasskey() {
    try {
      // get Registration Options from server
      const response_1 = await fetch(`${SERVER_HTTP_URL}/user/passkeys/begin-registration`, {
        credentials: "include",
        method: "GET",
      });

      const data = await response_1.json(); // assuming the response is in JSON format
      console.log("data", data); // log the data received from the server Options + session

      // modify to require residentKey = true
      data.Options.publicKey.authenticatorSelection.requireResidentKey = true;
      // pass creationOptions to authenticator to create Passkey (user verification usw.)
      const creationOptions = await data.Options.publicKey; // type it?
      const attResp = await startRegistration(await creationOptions);
      console.log("attResp", attResp);

      // pass signed Challenge + pubkey to Server
      const response_2 = await fetch(`${SERVER_HTTP_URL}/user/passkeys/finish-registration`, {
        method: "POST",
        body: JSON.stringify(attResp),
      });
      const data_2 = await response_2.json();
      console.log(data_2);
    } catch (error) {
      console.error(error);
    }
  }

  async function loginWithPasskey(autofill = true) {
    try {
      // get login options from my RP (challenge, ...)
      const response = await fetch(`${SERVER_HTTP_URL}/passkeys`, {
        method: "GET",
      });
      const options = await response.json();
      console.log("loginOptions", options);

      // let Authenticator sign challenge with stored pubKey (User Verification needed)
      const asseResp = await startAuthentication(options.publicKey, autofill); // true for autofill // if false its not asking which passkey to use, is ok?

      // post the response (signed challenge) to rp
      const verificationResp = await fetch(`${SERVER_HTTP_URL}/passkeys`, {
        method: "POST",
        body: JSON.stringify(asseResp),
      });
      console.log(await verificationResp.json());

      // TODO: check and handle verificationResp
    } catch (error) {
      console.log(error);
    }
  }

  // async function handleCreateAccount() {
  //   //create acc
  //   //create pk
  // }

  // https://dribbble.com/shots/7757250-Sign-up-revamp
  return (
    <div className="login-board">
      <div className="login-board__dialog">
        <div className="login-board__form-wrapper">
          <div className="login-board__form">
            <Link to="/">
              <ScrumlrLogo className="login-board__logo" accentColorClassNames={["accent-color--blue", "accent-color--purple", "accent-color--lilac", "accent-color--pink"]} />
            </Link>

            <h1>{t("LoginBoard.title")}</h1>

            <div className="login-board__passkey">
              <button style={{margin: "40px"}} onClick={() => registerPasskey()}>
                Create new passkey
              </button>
              <button onClick={() => loginWithPasskey(true)}>auth me</button>

              {/* <TextInput
                id="login-board__passkey"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                onKeyDown={(e: React.KeyboardEvent<HTMLInputElement>) => {
                  if (e.key === "Enter") {
                    handleCreateAccount();
                  }
                }}
                name="username" // needs to be username
                autoComplete="webauthn" //username email pw ?
                placeholder="Email or Passkey"
                aria-invalid={!email}
                loginType="passkeys"
                actions={
                  <TextInputAction title={"Sign up with passkey"} onClick={() => handleCreateAccount()}>
                    Continue
                  </TextInputAction>
                }
              /> */}
            </div>

            <hr className="login-board__divider" data-label="or" />

            <Button className="login-board__passkey-button" rightIcon={<KeyIcon />} onClick={() => loginWithPasskey(false)}>
              Sign in with a passkey
            </Button>

            <hr className="login-board__divider" data-label="or" />

            <LoginProviders originURL={`${window.location.origin}${redirectPath}`} />

            <hr className="login-board__divider" data-label="or" />

            <fieldset className="login-board__fieldset">
              <legend className="login-board__fieldset-legend">{t("LoginBoard.anonymousLogin")}</legend>

              <div className="login-board__username">
                <TextInputLabel label={t("LoginBoard.username")} htmlFor="login-board__username" />
                <TextInput
                  id="login-board__username"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  onKeyDown={(e: React.KeyboardEvent<HTMLInputElement>) => {
                    if (e.key === "Enter") {
                      handleLogin();
                    }
                  }}
                  maxLength={20}
                  aria-invalid={!displayName}
                  actions={
                    <TextInputAction title={t("LoginBoard.generateRandomName")} onClick={() => setDisplayName(getRandomName())}>
                      <RefreshIcon />
                    </TextInputAction>
                  }
                />
              </div>
              {!displayName && <ValidationError>{t("LoginBoard.usernameValidationError")}</ValidationError>}

              {SHOW_LEGAL_DOCUMENTS && (
                <label className="login-board__form-element login-board__terms">
                  <input type="checkbox" className="login-board__checkbox" defaultChecked={termsAccepted} onChange={() => setTermsAccepted(!termsAccepted)} />
                  <span className="login-board__terms-label">
                    <Trans
                      i18nKey="LoginBoard.acceptTerms"
                      components={{
                        terms: <Link to="/legal/termsAndConditions" target="_blank" />,
                        privacy: <Link to="/legal/privacyPolicy" target="_blank" />,
                      }}
                    />
                  </span>
                </label>
              )}
            </fieldset>
            {submitted && !termsAccepted && <ValidationError>{t("LoginBoard.termsValidationError")}</ValidationError>}

            <Button className="login-board__anonymous-login-button" color="primary" onClick={handleLogin}>
              {t("LoginBoard.login")}
            </Button>
          </div>
        </div>

        <HeroIllustration className="login-board__illustration" />
      </div>
    </div>
  );
};
