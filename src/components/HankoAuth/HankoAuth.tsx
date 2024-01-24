import {useLocation, useNavigate} from "react-router";
import { Auth } from "utils/auth";
import { Toast } from "utils/Toast";

// const tenantId = "64284d4b-750b-4c6b-a809-9601c6cd6ae4";

// const passkeyApi = tenant({
//   //get from process
//   tenantId: tenantId,
//   apiKey: "81EX6eV-rysIp2t7m8ZxYoJxNf2oJ9W2e5w_TW84qOJZ55YYWxRCuMj6Xl03BmuU8CFDbiP-yzOTmx_2IgmqWA==",
// });

interface State {
  from: {pathname: string};
}

export const HankoAuth = () => {
  const navigate = useNavigate();
  const location_ = useLocation();

  // TODO: ...
  let redirectPath = "/";
  if (location_.state) {
    redirectPath = (location_.state as State).from.pathname;
  }

  // passkey sign in and redirection to board path that is in history
  async function handlePasskeyLogin() {
    if (true) { //TODO terms accepted
      try {
        const succeeded = await Auth.signInWithPasskey();
        succeeded && navigate(redirectPath);
      } catch (err) {
        Toast.error({title: "LoginBoard.errorOnRedirect"});
      }
    }
    // setSubmitted(true);
  }

  // TODO: Mediation: conditional
  // useEffect(() => {
  //   loginWithPasskey();
  // }, [])

  return <button onClick={() => handlePasskeyLogin()}>Sign In With PK</button>;
};

export default HankoAuth;
