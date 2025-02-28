import { useCallback } from "react";
import {
  create,
  parseCreationOptionsFromJSON,
  get,
  parseRequestOptionsFromJSON,
} from "@github/webauthn-json/browser-ponyfill";

import "./App.css";
import { useNavigate } from "react-router";

const App: React.FC = () => {
  const navigate = useNavigate();

  const registerUser = useCallback(async (data: FormData) => {
    // パスキーがサポートされた環境かどうかを確認
    //
    // TODO: そもそもコンポーネントを表示しないなどの対応にする
    if (
      window.PublicKeyCredential &&
      PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable &&
      PublicKeyCredential.isConditionalMediationAvailable
    ) {
      Promise.all([
        PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable(),
        PublicKeyCredential.isConditionalMediationAvailable(),
      ]).then((results) => {
        if (results.every((result) => result !== true)) {
          alert("Passkey is not supported");

          return;
        }
      });
    }

    const username = data.get("username") as string;
    if (username === "") {
      alert("Please enter a username");

      return;
    }

    const optionsAPIRes = await fetch(
      `http://localhost:8080/registration/options`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          username: username,
        }),
      }
    );
    if (!optionsAPIRes.ok) {
      alert("Failed to get registration options");

      return;
    }
    const optJson = await optionsAPIRes.json();

    // TODO: `PublicKeyCredential.parseCreationOptionsFromJSON()`で置き換える
    //       https://developer.mozilla.org/en-US/docs/Web/API/PublicKeyCredential/parseCreationOptionsFromJSON_static
    const options = parseCreationOptionsFromJSON(optJson);
    const publicKeyCredential = await create(options);

    const verificationsAPIRes = await fetch(
      `http://localhost:8080/registration/verifications`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify(publicKeyCredential),
      }
    );
    if (!verificationsAPIRes.ok) {
      alert("Failed to verify registration");

      return;
    }

    alert("Successfully registered!");
  }, []);

  const login = useCallback(async () => {
    // パスキーがサポートされた環境かどうかを確認
    //
    // TODO: そもそもコンポーネントを表示しないなどの対応にする
    if (
      window.PublicKeyCredential &&
      PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable &&
      PublicKeyCredential.isConditionalMediationAvailable
    ) {
      Promise.all([
        PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable(),
        PublicKeyCredential.isConditionalMediationAvailable(),
      ]).then((results) => {
        if (results.every((result) => result !== true)) {
          alert("Passkey is not supported");

          return;
        }
      });
    }

    // NOTE: ユーザーネームレス認証を目指すので、ここは入力しなくていい
    // const username = data.get("username") as string;
    // if (username === "") {
    //   alert("Please enter a username");

    //   return;
    // }

    const optionsAPIResponse = await fetch(
      "http://localhost:8080/authentication/options",
      {
        method: "POST",
        credentials: "include",
      }
    );
    if (!optionsAPIResponse.ok) {
      alert("Failed to get registration options");

      return;
    }
    const optionsJSON = await optionsAPIResponse.json();

    // TODO: `PublicKeyCredential.parseRequestOptionsFromJSON()`で置き換える
    //       https://developer.mozilla.org/en-US/docs/Web/API/PublicKeyCredential/parseRequestOptionsFromJSON_static
    const options = parseRequestOptionsFromJSON(optionsJSON);
    const assertion = await get(options);

    const verificationsAPIResponse = await fetch(
      "http://localhost:8080/authentication/verifications",
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify(assertion),
      }
    );
    if (!verificationsAPIResponse.ok) {
      alert("Failed to verify registration");

      return;
    }
    const verificationsResultJSON = await verificationsAPIResponse.json();
    const userID = verificationsResultJSON.user_id;
    if (!userID) {
      alert("Failed to get user ID");

      return;
    }

    alert("Successfully logged in!");

    navigate(`/users/${userID}/public_keys`);
  }, [navigate]);

  return (
    <>
      <form>
        <div>
          <label>Username: </label>
          <input type="text" name="username" placeholder="i.e. foo@bar.com" />
        </div>
        <div>
          <label>Password: </label>
          <input type="password" name="password" placeholder="i.e. password" />
        </div>
        <br />
        <div>
          <button formAction={registerUser}>Register</button>
          <button formAction={login}>Login</button>
        </div>
      </form>
    </>
  );
};

export default App;
