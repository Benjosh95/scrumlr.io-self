import {useState} from "react";
import store, {useAppSelector} from "store";
import {Auth} from "utils/auth";
import {Toast} from "utils/Toast";
import {Actions} from "store/action";
import {ReactComponent as PasskeysIcon} from "assets/icon-passkey.svg";
import {ReactComponent as AddIcon} from "assets/icon-add.svg";
import {SettingsAccordion} from "./SettingsAccordion";
import "./PasskeySettings.scss";
import classNames from "classnames";

const PasskeySettings = () => {
  // const {t} = useTranslation();
  const [openAccordions, setOpenAccordions] = useState<{[key: string]: boolean}>({});
  const [credentialStates, setCredentialStates] = useState<{[key: string]: {showDeleteConfirmation: boolean; showRenameInput: boolean}}>({});
  const [renameValue, setRenameValue] = useState<string>("");

  const state = useAppSelector((applicationState) => ({
    user: applicationState.auth.user,
  }));

  const handleRegisterPasskey = async () => {
    await Auth.registerNewPasskey();
  }

  const handleDeletePasskey = (id: string) => {
    try {
      const updatedCredentials = state.user!.credentials!.filter((c) => c.ID !== id);
      store.dispatch(Actions.editSelf({...state.user!, credentials: updatedCredentials}, true));
      Toast.success({title: "Successfully deleted passkey"});
    } catch (error) {
      Toast.error({title: "Passkey could not be deleted"});
    }
  }

  const handleRenamePasskey = (id: string) => {
    try {
      const updatedCredentials = state.user?.credentials?.map((c) => {
        if (c.ID === id) {
          return {...c, displayName: renameValue};
        }
        return c;
      });
      store.dispatch(Actions.editSelf({...state.user!, credentials: updatedCredentials}, true));
      Toast.success({title: "Successfully renamed passkey"});
    } catch (error) {
      Toast.error({title: "Passkey could not be renamed"});
    } finally {
      setCredentialStates((prevState) => ({
        ...prevState,
        [id]: {...prevState[id], showRenameInput: !prevState[id]?.showRenameInput},
      }));
      setRenameValue("");
    }
  }

  const handleClickAccordion = (accordionId: string) => {
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

  const formatDateTime = (date: any) => {
    if (!date) {
      return "";
    }
    const dateObject = date instanceof Date ? date : new Date(date);
    return new Intl.DateTimeFormat("en-GB", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    }).format(dateObject);
  };

  return (
    <div className={classNames("profile-settings__passkey-container", "accent-color__lean-lilac")}>
      <div className="profile-settings__passkey-header">
        Passkeys Verwalten
        <PasskeysIcon className="profile-settings__passkey-header-icon" />
      </div>
      <div className="profile-settings__passkey">
        {state.user?.credentials?.map((credential) => {
          return (
            <SettingsAccordion
              key={credential.ID}
              isOpen={openAccordions[credential.ID]}
              onClick={() => handleClickAccordion(credential.ID)}
              label={credential.displayName ?? buildPasskeyId(credential.ID, 8)}
              headerClassName="avatar-settings__settings-group-header"
            >
              <ul className="profile-settings__passkey-controls">
                <li className="profile-settings__passkey-controls-rename">
                  <p>Rename passkey</p>
                  <p>Set a name for the passkey.</p>
                  <div
                  //TODO: fix weird rename input behavior
                  // onBlur={() => {
                  //   setCredentialStates((prevState) => ({
                  //     ...prevState,
                  //     [credential.ID]: {...prevState[credential.ID], showRenameInput: false},
                  //   }));
                  // }}
                  >
                    <button
                      onClick={() => {
                        setCredentialStates((prevState) => ({
                          ...prevState,
                          [credential.ID]: {...prevState[credential.ID], showRenameInput: !prevState[credential.ID]?.showRenameInput},
                        }));
                      }}
                    >
                      Rename
                    </button>
                    {credentialStates[credential.ID]?.showRenameInput && (
                      <form
                        onKeyDown={(e) => {
                          if (e.key === "Enter") {
                            e.preventDefault();
                            handleRenamePasskey(credential.ID);
                          }
                        }}
                      >
                        <input
                          ref={(input) => input?.focus()}
                          type="text"
                          id="renameInput"
                          autoComplete="off"
                          placeholder={renameValue}
                          value={renameValue}
                          onChange={(e) => setRenameValue(e.target.value)}
                        ></input>
                      </form>
                    )}
                  </div>
                </li>
                <li className="profile-settings__passkey-controls-delete">
                  <p>Delete passkey</p>
                  <p>Delete this passkey from your account.</p>
                  <div
                    onBlur={() => {
                      setCredentialStates((prevState) => ({
                        ...prevState,
                        [credential.ID]: {...prevState[credential.ID], showDeleteConfirmation: false},
                      }));
                    }}
                  >
                    <button
                      onClick={() => {
                        setCredentialStates((prevState) => ({
                          ...prevState,
                          [credential.ID]: {...prevState[credential.ID], showDeleteConfirmation: !prevState[credential.ID]?.showDeleteConfirmation},
                        }));
                      }}
                    >
                      Delete
                    </button>
                    {credentialStates[credential.ID]?.showDeleteConfirmation && (
                      <>
                        <button
                          onClick={() => {
                            setCredentialStates((prevState) => ({
                              ...prevState,
                              [credential.ID]: {...prevState[credential.ID], showDeleteConfirmation: !prevState[credential.ID]},
                            }));
                          }}
                        >
                          no
                        </button>
                        /<button onMouseDown={() => handleDeletePasskey(credential.ID)}>yes</button>
                      </>
                    )}
                  </div>
                </li>
                <li className="profile-settings__passkey-controls-last-used">
                  <p>Last used at</p>
                  <p>{formatDateTime(credential.lastUsedAt)}</p>
                </li>
                <li className="profile-settings__passkey-controls-created">
                  <p>Created at</p>
                  <p>{formatDateTime(credential.createdAt)}</p>
                </li>
              </ul>
            </SettingsAccordion>
          );
        })}
      </div>
      <button className="profile-settings__passkey-create-button" onClick={() => handleRegisterPasskey()}>
        Create a passkey
        <AddIcon className="profile-settings__passkey-create-button-icon" />
      </button>
    </div>
  );
};

export default PasskeySettings;
