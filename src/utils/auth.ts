import {Toast} from "utils/Toast";
import {API} from "api";
import i18n from "i18next";
import store from "store";
import {Actions} from "store/action";
import {SERVER_HTTP_URL} from "../config";
import {startAuthentication} from "@simplewebauthn/browser";

/**
 * Sign in anonymously.
 *
 * @param displayName Display name of the parse auth user.
 *
 * @returns Promise with user credentials on successful sign in, null otherwise.
 */
const signInAnonymously = async (displayName: string) => {
  try {
    const user = await API.signInAnonymously(displayName);
    if (user) {
      store.dispatch(Actions.signIn(user.id, user.name));
    }
    return true;
  } catch (err) {
    Toast.error({title: i18n.t("Toast.authenticationError")});
    return null;
  }
};

/**
 * Redirects to OAuth page of provider:
 *    https://accounts.google.com/o/oauth2/v2/auth/oauthchooseaccount...
 *    https://github.com/login/oauth/authorize...
 *    https://login.microsoftonline.com/common/oauth2/v2.0/authorize...
 * @param authProvider name of the OAuth Provider
 * @param originURL origin URL
 */
const signInWithAuthProvider = async (authProvider: string, originURL: string) => {
  window.location.href = `${SERVER_HTTP_URL}/login/${authProvider}?state=${encodeURIComponent(originURL)}`;
};

/**
 * Sign in with a passkey.
 *
 * @param autofill determines if passkey is used via conditional-UI
 *
 * @returns Promise with user credentials on successful sign in, null otherwise.
 */
const signInWithPasskey = async (autofill: boolean) => {
  try {
    // get login options from my RP (challenge, ...)
    const options = await API.getLoginOptions();
    console.log("loginOptions", options);

    // let the (platform-)Authenticator sign the challenge with on device stored pubKey
    const assertionResponse = await startAuthentication(options.publicKey, autofill); // This function triggers "User Verification" if false its not asking which passkey to use, is ok?
    console.log("assertionResponse", assertionResponse);

    // take assertionResponse (signed challenge, ...) from Authenticator and send it to  to RP
    const user = await API.verifyLogin(assertionResponse); //return user and check status code instead of "login success"
    console.log("verificationResp", user);

    if (user) {
      store.dispatch(Actions.signIn(user.id, user.name, undefined, user.credentials));
      return true;
    }

    throw new Error(`Could not sign in with passkey`);
  } catch (error) {
    Toast.error({title: i18n.t("Toast.authenticationError")});
    return null;
  }
};

const registerNewPasskey = async () => {
  try {
    const credentials = await API.registerNewPasskey();
    const user = await API.getCurrentUser();
    if (!credentials || !user) {
      throw new Error();
    }
    store.dispatch(Actions.editSelf({...user, credentials: credentials}, true));
    Toast.success({title: "Successfully registered new passkey"});
  } catch (error) {
    Toast.error({title: "Could not register new passkey"});
  }
};

export const Auth = {
  signInAnonymously,
  signInWithAuthProvider,
  signInWithPasskey,
  registerNewPasskey,
};
