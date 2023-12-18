import {useState} from "react";
import {SERVER_HTTP_URL} from "config";
import {startRegistration} from "@simplewebauthn/browser";
import {SettingsAccordion} from "./SettingsAccordion";

const PasskeySettings = () => {
  const [openPasskeyAccordion, setOpenPasskeyAccordion] = useState(false);

  async function registerPasskey() {
    try {
      // get registration options from server
      let response = await fetch(`${SERVER_HTTP_URL}/user/passkeys/begin-registration`, {
        method: "GET",
        credentials: "include",
      });

      const registrationOptions = await response.json();
      console.log("registrationOptions", registrationOptions);

      // modify to require residentKey = true
      // needed?
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

      const verificationResponse = await response.json();
      console.log("verificationResponse", verificationResponse);
    } catch (error) {
      console.error(error);
    }
  }

  return (
    <div className="profile-settings__passkey">
      <SettingsAccordion isOpen={openPasskeyAccordion} onClick={() => setOpenPasskeyAccordion((prevState) => !prevState)} label="Passkeys verwalten">
        <ul>Passkey-49dt39fk34</ul>
        <p>info</p>
        <p>rename</p>
        <p>delete</p>
        <ul>Passkey-49dt39fk34</ul>
        <button onClick={() => registerPasskey()}>+ Create a Passkey</button>
      </SettingsAccordion>
    </div>
  );
};

export default PasskeySettings;
