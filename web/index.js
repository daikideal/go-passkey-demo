/**
 * 以下より拝借
 *
 * https://github.com/subkaitaku/webauthn-example/blob/main/templates/index.js
 */

/**
 * URLBase64 to ArrayBuffer(Uint8Array)
 *
 * NOTE: go-webauthnの BeginRegistration で生成される challenge はURLBase64でエンコードされている。
 *       公式によると json.Unmarshal すればURLデコードしてくれるらしいのだが、2023-11-12現在、自分の環境だとそういう挙動にはなっていなかった。
 *       ただ、URLBase64で渡されることをクライアントと合意していれば問題ないと思われるので、デコード(= "-"を"+"に、"_"を"/"に置換)してから ArrayBuffer に変換する。
 *
 * @see {@link https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.8.6/protocol#URLEncodedBase64}
 * @see {@link https://qiita.com/kunihiros/items/2722d690b1525813c45e#base64-url}
 */
const bufferDecode = (value) => {
  return Uint8Array.from(
    atob(value.replace(/-/g, "+").replace(/_/g, "/")),
    (c) => c.charCodeAt(0)
  );
};

// ArrayBuffer(Uint8Array) to URLBase64
const bufferEncode = (value) =>
  btoa(String.fromCharCode(...new Uint8Array(value)))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");

const registerUser = () => {
  const username = document.getElementById("email").value;
  if (username === "") {
    alert("Please enter a username");
    return;
  }

  fetch(`/registration/options`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      username: username,
    }),
  })
    .then((response) => response.json())
    .then((credentialCreationOptions) => {
      credentialCreationOptions.publicKey.challenge = bufferDecode(
        credentialCreationOptions.publicKey.challenge
      );
      credentialCreationOptions.publicKey.user.id = bufferDecode(
        credentialCreationOptions.publicKey.user.id
      );
      if (credentialCreationOptions.publicKey.excludeCredentials) {
        credentialCreationOptions.publicKey.excludeCredentials.forEach(
          (item) => {
            item.id = bufferDecode(item.id);
          }
        );
      }

      return navigator.credentials.create({
        publicKey: credentialCreationOptions.publicKey,
      });
    })
    .then((credential) => {
      const attestationObject = credential.response.attestationObject;
      const clientDataJSON = credential.response.clientDataJSON;
      const rawId = credential.rawId;

      // 認証機の登録を完了
      return fetch(`/registration/verifications`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          id: credential.id,
          rawId: bufferEncode(rawId),
          type: credential.type,
          response: {
            attestationObject: bufferEncode(attestationObject),
            clientDataJSON: bufferEncode(clientDataJSON),
          },
          username: username,
        }),
      });
    })
    .then((response) => response.json())
    .then(() => {
      alert("Successfully registered " + username + "!");
    })
    .catch((error) => {
      console.log(error);
      alert("Failed to register " + username);
    });
};

const loginUser = () => {
  // NOTE: ユーザーネームレス認証を目指すので、ここは入力しなくていい
  //   const username = document.getElementById("email").value;
  //   if (username === "") {
  //     alert("Please enter a username");
  //     return;
  //   }

  fetch("/authentication/options", { method: "POST" })
    .then((response) => response.json())
    .then((credentialRequestOptions) => {
      credentialRequestOptions.publicKey.challenge = bufferDecode(
        credentialRequestOptions.publicKey.challenge
      );
      // NOTE: allowCredentials には 構造体タグで omitempty がついており、
      //       レスポンスにそもそも含まれていなかったりするのでここでケアする。
      credentialRequestOptions.publicKey.allowCredentials
        ? credentialRequestOptions.publicKey.allowCredentials.forEach(
            (item) => {
              item.id = bufferDecode(item.id);
            }
          )
        : (credentialRequestOptions.publicKey.allowCredentials = undefined);

      return navigator.credentials.get({
        publicKey: credentialRequestOptions.publicKey,
      });
    })
    .then((assertion) => {
      console.log("assertion", { assertion });

      const authData = assertion.response.authenticatorData;
      const clientDataJSON = assertion.response.clientDataJSON;
      const rawId = assertion.rawId;
      const sig = assertion.response.signature;
      const userHandle = assertion.response.userHandle;

      return fetch("/authentication/verifications", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          id: assertion.id,
          rawId: bufferEncode(rawId),
          type: assertion.type,
          response: {
            authenticatorData: bufferEncode(authData),
            clientDataJSON: bufferEncode(clientDataJSON),
            signature: bufferEncode(sig),
            userHandle: bufferEncode(userHandle),
          },
        }),
      });
    })
    .then((response) => response.json())
    .then(() => {
      alert("Successfully logged in!");
    })
    .catch((error) => {
      console.error(error);
      alert("Failed to login!");
    });
};
