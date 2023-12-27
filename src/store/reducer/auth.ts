import {AuthState} from "types/auth";
import {ReduxAction} from "store/action";
import {AuthAction} from "store/action/auth";
import { ParticipantAction } from "store/action/participants";

// eslint-disable-next-line @typescript-eslint/default-param-last
export const authReducer = (state: AuthState = {user: undefined, initializationSucceeded: null}, action: ReduxAction): AuthState => {
  if (action.type === AuthAction.SignOut) {
    return {
      ...state,
      user: undefined,
    };
  }

  if (action.type === AuthAction.SignIn) {
    return {
      ...state,
      user: {
        id: action.id,
        name: action.name,
        avatar: action.avatar,
        credentials: action.credentials,
      },
    };
  }

  if (action.type === ParticipantAction.EditSelf){ //bad practice?
    if(!action.editCredentials) {
      return state
    }
    return {
      ...state,
      user: {
        ...action.user,
        credentials: action.user.credentials,
      }
    }
  }

  if (action.type === AuthAction.UserCheckCompleted) {
    return {
      ...state,
      initializationSucceeded: action.success,
    };
  }

  return state;
};
