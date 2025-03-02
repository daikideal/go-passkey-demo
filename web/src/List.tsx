import React, { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router";

import "./List.css";

type UserInfo = {
  id: string;
  name: string;
};

type PublicKey = {
  id: string;
  credential_id: string;
  public_key: string;
  attestation_type: string;
  transport: string[];
  flags: object;
  authenticator: object;
};

const PublicKeyTableColumns: (keyof PublicKey)[] = [
  "credential_id",
  "public_key",
  "attestation_type",
  "transport",
  "flags",
  "authenticator",
];

/**
 * TODO: 表示する内容を精査する。
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
            {PublicKeyTableColumns.map((key) => (
              <th key={key}>{key}</th>
            ))}
            <th>削除する</th>
          </tr>
        </thead>
        <tbody>
          {publicKeys.map((publicKey) => (
            <tr key={publicKey.id}>
              {PublicKeyTableColumns.map((key) => (
                <td key={key} headers={key}>
                  {
                    // NOTE: パスキー管理画面に何を表示したらいいのかわからないまま実装しているため、一旦全て表示しようとした結果、ArrayやObjectが混ざる。
                    //  - Array: カンマ区切り
                    //  - Object: JSON.stringify
                    //  - それ以外: そのまま表示
                    Array.isArray(publicKey[key])
                      ? publicKey[key].join(", ")
                      : typeof publicKey[key] === "object"
                      ? JSON.stringify(publicKey[key])
                      : publicKey[key]
                  }
                </td>
              ))}

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
