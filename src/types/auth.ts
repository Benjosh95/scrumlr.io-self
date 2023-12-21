import {AvataaarProps} from "components/Avatar";
import {Credential} from "./credential"
export interface Auth {
  id: string;
  name: string;
  avatar?: AvataaarProps;
  credentials?: Credential[]; // create type
}

export interface AuthState {
  user: Auth | undefined;
  initializationSucceeded: boolean | null;
}
