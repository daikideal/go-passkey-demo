import React, { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router";

import "./List.css";

type UserInfo = {
  id: string;
  name: string;
};

/**
 * https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.12.2/webauthn#Authenticator
 */
type Authenticator = {
  AAGUID: string;
  signCount: number;
  cloneWarning: boolean;
  attachment: string;
};

/**
 * https://pkg.go.dev/github.com/go-webauthn/webauthn@v0.12.2/webauthn#Credential
 */
type PublicKey = {
  id: string;
  credential_id: string;
  public_key: string;
  attestation_type: string;
  transport: string[];
  flags: object;
  authenticator: Authenticator;
};

/**
 * TODO: 表示する内容を精査する。
 * - パスキーの名前とアイコン...AAGUIDから特定 or UserAgentから特定
 * - 登録日時・最終使用日時・使用したOS
 * - 同期パスキーのラベル...BEフラグを見る
 * - 名前の編集ボタン
 * - 削除ボタン
 */
const PublicKeyList: React.FC = () => {
  const { id: userID } = useParams();

  const [userInfo, setUserInfo] = useState<UserInfo>();
  const [publicKeys, setPublicKeys] = useState<PublicKey[]>([]);

  // NOTE: 本当はuseEffectでデータフェッチしたくないけど、ライブラリ入れるのも面倒に感じたので一旦これで…。
  useEffect(() => {
    fetch(`http://localhost:8080/users/${userID}`)
      .then((res) => res.json())
      .then((json) => {
        setUserInfo(json);
      })
      .catch((err) => alert(err));

    fetch(`http://localhost:8080/users/${userID}/public_keys`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
    })
      .then((res) => res.json())
      .then((json) => {
        setPublicKeys(json);
      })
      .catch((err) => alert(err));
  }, [userID]);

  /**
   * Base64エンコードされたAAGUIDのバイナリを、UUIDフォーマットの文字列に変換する
   *
   * @example
   *    "utpVZqeqQB+9lkVhmlUSDQ==" => "bada5566-a7aa-401f-bd96-45619a55120d" (1Password)
   *    "rc4AAjW8xgpkiwsl8fBVAw==" => "adce0002-35bc-c60a-648b-0b25f1f05503" (Chrome on Mac)
   *
   * @see https://github.com/passkeydeveloper/passkey-authenticator-aaguids/blob/main/aaguid.json
   * @see https://github.com/web-auth/webauthn-framework/pull/49
   */
  const parseAaguidAsUuid = useCallback((base64AaguidBinary: string) => {
    const decoded = atob(base64AaguidBinary);
    const hexString = Array.from(decoded)
      .map((char) => char.charCodeAt(0).toString(16).padStart(2, "0"))
      .join("");

    return `${hexString.slice(0, 8)}-${hexString.slice(
      8,
      12
    )}-${hexString.slice(12, 16)}-${hexString.slice(16, 20)}-${hexString.slice(
      20
    )}`;
  }, []);

  const deletePublicKey = useCallback(
    async (id: string) => {
      const deleteAPIRes = await fetch(
        `http://localhost:8080/users/${userID}/public_keys/${id}`,
        {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
          },
          credentials: "include",
        }
      );
      if (!deleteAPIRes.ok) {
        alert(`Failed to delete public key: ${id}`);

        return;
      }

      setPublicKeys((prevPublicKeys) =>
        prevPublicKeys.filter((publicKey) => publicKey.id !== id)
      );
    },
    [userID]
  );

  return (
    <div>
      <h1>{userInfo?.name}'s passkeys</h1>
      <table>
        <caption>{userInfo?.name}'s passkeys</caption>
        <thead>
          <tr>
            <th key="aaguid">AAGUID</th>
            <th>削除する</th>
          </tr>
        </thead>
        <tbody>
          {publicKeys.map((publicKey) => (
            <tr key={publicKey.id}>
              <td key="aaguid">
                <span>{parseAaguidAsUuid(publicKey.authenticator.AAGUID)}</span>
              </td>
              <td>
                <button onClick={() => deletePublicKey(publicKey.id)}>
                  削除
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default PublicKeyList;
