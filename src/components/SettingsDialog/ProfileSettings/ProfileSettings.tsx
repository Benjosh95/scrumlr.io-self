import classNames from "classnames";
import {useTranslation} from "react-i18next";
import {useEffect, useState} from "react";
import store, {useAppSelector} from "store";
import {Actions} from "store/action";
import "./ProfileSettings.scss";
import {useDispatch} from "react-redux";
import {ReactComponent as InfoIcon} from "assets/icon-info.svg";
import {Toggle} from "components/Toggle";
import {SERVER_HTTP_URL} from "config";
import {AvatarSettings} from "../Components/AvatarSettings";
import {SettingsInput} from "../Components/SettingsInput";
import {SettingsButton} from "../Components/SettingsButton";
import "@corbado/webcomponent/pkg/auth_cui.css";
import "@corbado/webcomponent";

interface AssociationToken {
  associationToken: string;
}

export const ProfileSettings = () => {
  const {t} = useTranslation();
  const dispatch = useDispatch();

  const state = useAppSelector((applicationState) => ({
    participant: applicationState.participants!.self,
    hotkeysAreActive: applicationState.view.hotkeysAreActive,
  }));

  const [userName, setUserName] = useState<string>(state.participant?.user.name);
  const [id] = useState<string | undefined>(state.participant?.user.id);
  const [associationToken, setAssociationToken] = useState<AssociationToken | null>(null);

  // Instead of clicking on a button, you can also start the backend API call, when the user opens a new page
  // It's only important to note, that you can only use the <corbado-passkey-associate/> web component if
  // an association token has been created before.
  useEffect(() => {
    handleButtonClick();
  }, []);

  const handleButtonClick = async () => {
    try {
      // loginIdentifier needs to be obtained via a backend call or your current state / session management
      // it should be a dynamic value depending on the current logged-in user
      const response = await fetch(`${SERVER_HTTP_URL}/passkeys/createAssociationToken`, {
        method: "GET",
        credentials: "include",
      });

      const token = await response.json();
      console.log(token);

      // const response = await axios.post<AssociationToken>(process.env.NEXT_PUBLIC_API_BASE_URL + "/api/createAssociationToken", {
      //     loginIdentifier: "vincent+20@corbado.com",
      //     loginIdentifierType: "email"
      // })

      setAssociationToken({associationToken: token});
    } catch (err) {
      console.log(err);
    }
  };

  return (
    <div className={classNames("settings-dialog__container", "accent-color__lean-lilac")}>
      <header className="settings-dialog__header">
        <h2 className="settings-dialog__header-text">{t("ProfileSettings.Profile")}</h2>
      </header>
      <div className="profile-settings__container-wrapper">
        <div className="profile-settings__container">
          <SettingsInput
            id="profileSettingsUserName"
            label={t("ProfileSettings.UserName")}
            value={userName}
            onChange={(e) => setUserName(e.target.value)}
            submit={() => store.dispatch(Actions.editSelf({...state.participant.user, name: userName}))}
          />

          <AvatarSettings id={id} />

          <button onClick={handleButtonClick}>Add passkey to my account</button>

          {associationToken && <corbado-passkey-associate project-id="pro-1697688654308671815" association-token={associationToken.associationToken} />}

          <div className="profile-settings__hotkey-settings">
            <SettingsButton
              className="profile-settings__toggle-hotkeys-button"
              label={t("Hotkeys.hotkeyToggle")}
              onClick={() => {
                dispatch(Actions.setHotkeyState(!state.hotkeysAreActive));
              }}
            >
              <Toggle active={state.hotkeysAreActive} />
            </SettingsButton>
            <a className="profile-settings__open-cheat-sheet-button" href={`${process.env.PUBLIC_URL}/hotkeys.pdf`} target="_blank" rel="noopener noreferrer">
              <p>{t("Hotkeys.cheatSheet")}</p>
              <InfoIcon />
            </a>
          </div>
        </div>
      </div>
    </div>
  );
};
