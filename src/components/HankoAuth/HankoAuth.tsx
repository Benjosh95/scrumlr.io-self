import {tenant} from "@teamhanko/passkeys-sdk";
import {CredentialRequestOptionsJSON, get} from "@github/webauthn-json";
import {createRemoteJWKSet, jwtVerify} from "jose";
import Cookies from "js-cookie";

const tenantId = "64284d4b-750b-4c6b-a809-9601c6cd6ae4";

const passkeyApi = tenant({
  //get from process
  tenantId: tenantId, 
  apiKey: "81EX6eV-rysIp2t7m8ZxYoJxNf2oJ9W2e5w_TW84qOJZ55YYWxRCuMj6Xl03BmuU8CFDbiP-yzOTmx_2IgmqWA==",
});

export const HankoAuth = () => {
  const loginWithPasskey = async () => {
    const loginOptions = (await passkeyApi.login.initialize()) as CredentialRequestOptionsJSON;
    console.log("loginOptions", loginOptions);

    // const assertionResponse = await get({...loginOptions, mediation: 'conditional'}); //TODO: Mediation: conditional -> create corresponding input
    const assertionResponse = await get(loginOptions);
    console.log("assertionResponse", assertionResponse);

    // Should i set the JWT/Token by myself from component? wtf
    const verificationResponse = await passkeyApi.login.finalize(assertionResponse as any); // TODO: Ok?
    const token = verificationResponse.token;
    console.log("token", token);

    // TODO: not working right now?
    // maybe use backend endpoints instead of the sdk. This way i could set the cookie over the Response instead of from here
    if (token) {
      // const secureFlag = window.location.protocol === "https:" ? "Secure" : "";
      Cookies.set("hanko", token, {path:"/", expires: 1, sameSite: "strict", domain:`${location.hostname}`}) // TODO: httpOnly but isnt working
      // const readCookie = Cookies.get("hanko")
    }

    //get from process
    const JWKS = createRemoteJWKSet(new URL(`https://passkeys.hanko.io/${tenantId}/.well-known/jwks.json`));
    try {
      const verifiedJWT = await jwtVerify(token ?? "", JWKS);
      //TODO
      console.log("verifiedJWT", verifiedJWT);
    } catch {
      //TODO
    }
  };

  // const setCookie = (name: string, value: string, expirationDays: number) => {
  //   const expirationDate = new Date(Date.now() + expirationDays * 24 * 60 * 60 * 1000).toUTCString();
  //   const secureFlag = window.location.protocol === "https:" ? "Secure" : "";
  //   const cookieOptions = `path=/; expires=${expirationDate}; HttpOnly; ${secureFlag}; SameSite=Strict`;

  //   document.cookie = `${name}=${value}; ${cookieOptions}`;
  // };

  // TODO: Mediation: conditional
  // useEffect(() => {
  //   loginWithPasskey();
  // }, [])

  return <button onClick={() => loginWithPasskey()}>Sign In With PK</button>;
};

export default HankoAuth;
