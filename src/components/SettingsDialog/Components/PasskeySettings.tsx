import {useState} from "react";
import {SERVER_HTTP_URL} from "config";
import {startRegistration} from "@simplewebauthn/browser";
import {SettingsAccordion} from "./SettingsAccordion";
import { useAppSelector } from "store";

const PasskeySettings = () => {
  const [openPasskeyAccordion, setOpenPasskeyAccordion] = useState(false);

  const state = useAppSelector((applicationState) => ({
    user: applicationState.auth.user,
  }));
  
  //TODO: Refactor into middleware?
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
        {state.user?.credentials?.map((c) => {
          return <ul>{c.ID}</ul>
        })}
        {/* <ul>Passkey-49dt39fk34</ul>
        <p>info</p>
        <p>rename</p>
        <p>delete</p>
        <ul>Passkey-49dt39fk34</ul> */}
        <button onClick={() => registerPasskey()}>+ Create a Passkey</button>
      </SettingsAccordion>
    </div>
  );
};

export default PasskeySettings;
