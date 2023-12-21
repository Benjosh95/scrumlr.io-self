import {Auth} from "types/auth";
import {SERVER_HTTP_URL} from "../config";

export const UserAPI = {
  /**
   * Edits a user.
   *
   * @param user the updated user object
   * @returns a {status, description} object
   */
  editUser: async (user: Auth, editCredentials: boolean) => { //what a nasty workaround :D 
    console.log("editCredentials", editCredentials)
    try {
      if (!editCredentials) {
        let response = await fetch(`${SERVER_HTTP_URL}/user/`, {
          method: "GET",
          credentials: "include",
        });

        let currentUser: Auth;
        if (response.status === 200) {
          currentUser = (await response.json()) as Auth;
          user.credentials = currentUser.credentials;
          console.log("currentUser", currentUser)
        }

        response = await fetch(`${SERVER_HTTP_URL}/user/`, {
          method: "PUT",
          credentials: "include",
          body: JSON.stringify(user),
        });

        if (response.status === 200) {
          return (await response.json()) as Auth;
        }

        // console.log("currentUser", currentUser)

        throw new Error(`request resulted in response status ${response.status}`);
      } else {
        const response = await fetch(`${SERVER_HTTP_URL}/user/`, {
          method: "PUT",
          credentials: "include",
          body: JSON.stringify(user),
        });

        if (response.status === 200) {
          return (await response.json()) as Auth;
        }

        throw new Error(`request resulted in response status ${response.status}`);
      }
    } catch (error) {
      throw new Error(`unable to update user: ${error}`);
    }
  },
};
