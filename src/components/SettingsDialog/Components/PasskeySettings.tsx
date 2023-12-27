import {useState} from "react";
import store, {useAppSelector} from "store";
import {Auth} from "utils/auth";
import {Toast} from "utils/Toast";
import {useTranslation} from "react-i18next";
import {Actions} from "store/action";
import {ReactComponent as KeyIcon} from "assets/icon-key.svg";
import {ReactComponent as AddIcon} from "assets/icon-add.svg";
import {SettingsAccordion} from "./SettingsAccordion";
import "./PasskeySettings.scss";
import classNames from "classnames";

const PasskeySettings = () => {
  const {t} = useTranslation();
  const [openAccordions, setOpenAccordions] = useState<{[key: string]: boolean}>({});

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
      const updatedCredentials = state.user!.credentials!.filter((c) => c.ID !== id);
      store.dispatch(Actions.editSelf({...state.user!, credentials: updatedCredentials}, true));
    } catch (error) {
      Toast.error({title: t("Passkey could not be deleted")});
    }
  }

  // TODO: add field in DB befor implementin this
  // async function handlePasskeyRenaming() {
  //   try {
  //     console.log("trigger renaming");
  //   } catch (error) {
  //     Toast.error({title: t("Passkey could not be renamed")});
  //   }
  // }

  const handleAccordionClick = (accordionId: string) => {
    setOpenAccordions((prevState) => ({
      ...prevState,
      [accordionId]: !prevState[accordionId],
    }));
  };

  const buildPasskeyId = (str: string, maxLength: number) => {
    if (str.length > maxLength) {
      return `Passkey-${str.slice(0, maxLength)}`;
    }
    return str;
  };

  return (
    <div className={classNames("profile-settings__passkey-container", "accent-color__lean-lilac")}>
      <div className="profile-settings__passkey-header">
        Passkeys Verwalten
        <KeyIcon className="profile-settings__passkey-icon" />
      </div>
      <div className="profile-settings__passkey">
        {state.user?.credentials?.map((credential) => {
          return (
            //TODO: label = Condition if PasskeyName exists, else buildPasskeyID
            <SettingsAccordion
              key={credential.ID}
              isOpen={openAccordions[credential.ID]}
              onClick={() => handleAccordionClick(credential.ID)}
              label={buildPasskeyId(credential.ID, 8)}
            >
              <button onClick={() => handlePasskeyDeletion(credential.ID)}>Delete</button>
            </SettingsAccordion>
          );
        })}
      </div>
      <button className="profile-settings__passkey-create-button" onClick={() => handlePasskeyRegistration()}>
        Create a Passkey
        <AddIcon className="profile-settings__passkey-create-button-icon" />
      </button>
    </div>
  );
};

export default PasskeySettings;
