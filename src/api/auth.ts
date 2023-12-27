import { startRegistration } from "@simplewebauthn/browser";
import {SERVER_HTTP_URL} from "../config";
import {Auth} from "../types/auth";
import {AuthenticationResponseJSON} from "@simplewebauthn/typescript-types"

export const AuthAPI = {
  /**
   * Signs out the current user by deleting the session cookie.
   *
   * Since the session cookie is set to http only it cannot be accessed by the client. Therefore
   * a call to the server is required.
   */
  signOut: async () => {
    try {
      await fetch(`${SERVER_HTTP_URL}/login`, {
        method: "DELETE",
        credentials: "include",
      });
    } catch (error) {
      throw new Error(`unable to sign out with error: ${error}`);
    }
  },

  /**
   * Sign in by an anonymous account with the specified username.
   *
   * @param name the username of the account
   */
  signInAnonymously: async (name: string) => {
    try {
      const response = await fetch(`${SERVER_HTTP_URL}/login/anonymous`, {
        method: "POST",
        credentials: "include",
        body: JSON.stringify({name}),
      });

      if (response.status === 201) {
        const body = await response.json();
        return {
          id: body.id,
          name: body.name,
        };
      }

      throw new Error(`sign in request resulted in response status ${response.status}`);
    } catch (error) {
      throw new Error(`unable to sign in with error: ${error}`);
    }
  },

  /**
   * Registers a new Passkey for a user 
   * 
   * @returns all Credentials of the user, including the new one
   */
  registerNewPasskey: async () => {
    try {
      // get registration options from server
      let response = await fetch(`${SERVER_HTTP_URL}/user/passkeys/begin-registration`, {
        method: "GET",
        credentials: "include",
      });

      if(response.status !== 200) {
        throw new Error(`register passkey request resulted in response status ${response.status}`);
      }
      const registrationOptions = await response.json();
      console.log("registrationOptions", registrationOptions);
      // modify to require residentKey = true
      // registrationOptions.publicKey.authenticatorSelection.requireResidentKey = true

      // pass registration options to authenticator to create Passkey + user verification + sign challenge
      const authenticatorResponse = await startRegistration(await registrationOptions.publicKey);
      console.log("authenticatorResponse", authenticatorResponse);

      // pass signed Challenge + pubkey to Server
      response = await fetch(`${SERVER_HTTP_URL}/user/passkeys/finish-registration`, {
        method: "POST",
        credentials: "include",
        body: JSON.stringify(authenticatorResponse),
      });

      if(response.status === 200){
        const verificationResponse = await response.json();
        console.log("verificationResponse", verificationResponse);
        return verificationResponse;
      }
      throw new Error(`register passkey request resulted in response status ${response.status}`);
    } catch (error) {
      throw new Error(`unable to register new passkey: ${error}`);
    }
    return undefined;
  },

  /**
   * Gets options to login with passkey
   */
  getLoginOptions: async () => {
    try {
      // get login options from my RP (challenge, ...)
      const response = await fetch(`${SERVER_HTTP_URL}/login/passkeys/begin-authentication`, {
        method: "GET",
        credentials: "include",
      });

      if (response.status === 200) {
        const body = await response.json();
        return body;
      }

      throw new Error(`request to get login options resulted in response status ${response.status}`);
    } catch (error) {
      throw new Error(`unable to get login options in with error: ${error}`);
    }
  },

    /**
   * Verifies the chosen Passkey to login with passkey
   */
    verifyLogin: async (assertionResp: AuthenticationResponseJSON) => { 
      try {
        // post the response (signed challenge) from Authenticator to RP
          const response = await fetch(`${SERVER_HTTP_URL}/login/passkeys/finish-authentication`, {
            method: "POST",
            credentials: "include",
            body: JSON.stringify(assertionResp),
          });

        if (response.status === 200) {
          const body = await response.json();
          return body;
        }
        
        throw new Error(`request to verify login resulted in response status ${response.status}`);
      } catch (error) {
        throw new Error(`unable to verify login with error: ${error}`);
      }
    },

  /**
   * Returns the current user or `undefined`, if no session is available.
   *
   * @returns the user or `undefined`
   */
  getCurrentUser: async () => {
    try {
      const response = await fetch(`${SERVER_HTTP_URL}/user`, {
        method: "GET",
        credentials: "include",
      });

      if (response.status === 200) {
        return (await response.json()) as Auth;
      }
    } catch (error) {
      throw new Error(`unable to fetch current user: ${error}`);
    }

    return undefined;
  },
};
