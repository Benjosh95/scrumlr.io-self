import {Toast} from "utils/Toast";
import {API} from "api";
import i18n from "i18next";
import store from "store";
import {Actions} from "store/action";
import {SERVER_HTTP_URL} from "../config";
import { startAuthentication } from "@simplewebauthn/browser";

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
    const options = await API.getLoginOptions()
    console.log("loginOptions", options);

    // let the Authenticator sign the challenge with device stored pubKey (User Verification needed)
    const assertionResp = await startAuthentication(options.publicKey, autofill); // if false its not asking which passkey to use, is ok?
    console.log("assertionResp", assertionResp)

    // post the response (signed challenge) from Authenticator to RP
    const user = await API.verifyLogin(assertionResp) //return user and check status code instead of "login success"
    console.log("verificationResp", user)

    if (user) {
      store.dispatch(Actions.signIn(user.id, user.name, undefined, user.credentials));
    }
    return true;
    // throw new Error(`Could not sign in with passkey`);
  } catch (error) {
    console.log(error);
    return null
  }
};

const registerNewPasskey = async () => {
  try {
    const credentials = await API.registerNewPasskey();
    const user = await API.getCurrentUser();
    if(credentials && user) {
      //CONTINUE
      //alle neuen Credentials kommen, werden aber nicht korrekt bei editself gesetzt anscheinend
      console.log("credentials = ", credentials);
      store.dispatch(Actions.editSelf({...user, credentials: credentials}, true));
      return true;
    }
    return undefined;
  } catch (error) {
    console.error(error);
    return null //better undefined?
  }
}

export const Auth = {
  signInAnonymously,
  signInWithAuthProvider,
  registerNewPasskey,
  signInWithPasskey
};
