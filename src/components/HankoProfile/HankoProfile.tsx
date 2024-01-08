import {tenant} from "@teamhanko/passkeys-sdk";
import {create, CredentialCreationOptionsJSON} from "@github/webauthn-json";
import {SERVER_HTTP_URL} from "config";
import {useAppSelector} from "store";

const passkeyApi = tenant({
  tenantId: "64284d4b-750b-4c6b-a809-9601c6cd6ae4",
  apiKey: "81EX6eV-rysIp2t7m8ZxYoJxNf2oJ9W2e5w_TW84qOJZ55YYWxRCuMj6Xl03BmuU8CFDbiP-yzOTmx_2IgmqWA==",
});

export const HankoProfile = () => {
  const state = useAppSelector((applicationState) => ({
    user: applicationState.auth.user!,
    hotkeysAreActive: applicationState.view.hotkeysAreActive,
  }));

  async function registerPasskey() {
    const creationOptions = await passkeyApi.registration.initialize({
      userId: state.user.id,
      username: state.user.name,
    });

    const credential = await create(creationOptions as CredentialCreationOptionsJSON);

    const response = await fetch(`${SERVER_HTTP_URL}/passkey/finalize-registration`, {
      method: "POST",
      credentials: "include",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(credential),
    });
    const data = await response.json();
    console.log(data.token);
  }

  return <button onClick={registerPasskey}>Register Passkey</button>;
};

export default HankoProfile;
