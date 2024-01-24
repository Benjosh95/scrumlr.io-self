import {Toast} from "utils/Toast";
import {API} from "api";
import i18n from "i18next";
import store from "store";
import {Actions} from "store/action";
import {SERVER_HTTP_URL} from "../config";
import {tenant} from "@teamhanko/passkeys-sdk";
import {CredentialRequestOptionsJSON, get} from "@github/webauthn-json";

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
  TODO
 */
const signInWithPasskey = async () => {
  // TODO: passkeyApi Ok here? in try/catch?
  const passkeyApi = tenant({
    //TODO: get from process
    tenantId: "64284d4b-750b-4c6b-a809-9601c6cd6ae4",
    apiKey: "81EX6eV-rysIp2t7m8ZxYoJxNf2oJ9W2e5w_TW84qOJZ55YYWxRCuMj6Xl03BmuU8CFDbiP-yzOTmx_2IgmqWA==",
  });

  try {
    const loginOptions = (await passkeyApi.login.initialize()) as CredentialRequestOptionsJSON;
    console.log("loginOptions", loginOptions);

    // const assertionResponse = await get({...loginOptions, mediation: 'conditional'}); //TODO: Mediation: conditional -> create corresponding input
    const assertionResponse = await get(loginOptions);
    console.log(assertionResponse);

    const user = await API.signInWithPasskey(assertionResponse);

    if (user) {
      store.dispatch(Actions.signIn(user.id, user.name));
    }
    return true;
  } catch (err) {
    Toast.error({title: i18n.t("Toast.authenticationError")});
    return null;
  }
};

export const Auth = {
  signInAnonymously,
  signInWithAuthProvider,
  signInWithPasskey,
};
