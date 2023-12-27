import {useState} from "react";
import {SettingsAccordion} from "./SettingsAccordion";
import store, {useAppSelector} from "store";
import {Auth} from "utils/auth";
import {Toast} from "utils/Toast";
import {useTranslation} from "react-i18next";
import { Actions } from "store/action";

const PasskeySettings = () => {
  const {t} = useTranslation();
  const [openPasskeyAccordion, setOpenPasskeyAccordion] = useState(false);

  const state = useAppSelector((applicationState) => ({
    user: applicationState.auth.user,
  }));

  async function handlePasskeyRegistration() {
    try {
      const success = await Auth.registerNewPasskey();

      //show success anyhow?
      if (success) {
        console.log("successfully registered");
      }
    } catch (error) {
      Toast.error({title: t("Passkey could not be registered")});
    }
  }

  async function handlePasskeyDeletion(id: string) {
    try {
        const updatedCredentials = state.user!.credentials!.filter((c) => c.ID !== id) 
        store.dispatch(Actions.editSelf({...state.user!, credentials: updatedCredentials}, true))
    } catch (error) {
      Toast.error({title: t("Passkey could not be deleted")});
    }
  }  
  async function handlePasskeyRenaming() {
    try {
      console.log("trigger renaming")
    } catch (error) {
      Toast.error({title: t("Passkey could not be renamed")});
    }
  }
  
  return (
    <div className="profile-settings__passkey">
      <SettingsAccordion isOpen={openPasskeyAccordion} onClick={() => setOpenPasskeyAccordion((prevState) => !prevState)} label="Passkeys verwalten">
        {state.user?.credentials?.map((credential, idx) => {
          return (
            <ul key={credential.ID}>
              {idx + 1} {credential.ID}
              <button onClick={() => handlePasskeyDeletion(credential.ID)}>Delete</button>
              <button onClick={() => handlePasskeyRenaming()}>Rename</button>
            </ul>
          );
          //rename, delete, lastused at createdat
        })}
        <button onClick={() => handlePasskeyRegistration()}>+ Create a Passkey</button>
        {/* <button onClick={() => renamePasskey()}>? Rename</button> */}
        {/* <button onClick={() => deletePasskey()}>! Delete</button> */}
      </SettingsAccordion>
    </div>
  );
};

export default PasskeySettings;
