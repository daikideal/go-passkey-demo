import React, { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router";

import "./List.css";

type UserInfo = {
  id: string;
  name: string;
};

type PasskeyInfo = {
  id: string;
  AAGUID: string;
};

/**
 * @see https://github.com/passkeydeveloper/passkey-authenticator-aaguids/blob/main/aaguid.json
 */
type Aaguid = Record<
  string,
  {
    name: string;
    icon_dark: string | undefined;
    icon_light: string | undefined;
  }
>;

/**
 * TODO: 表示する内容を精査する。
 * - パスキーの名前とアイコン...AAGUIDから特定 or UserAgentから特定
 * - 登録日時・最終使用日時・使用したOS
 * - 同期パスキーのラベル...BEフラグを見る
 * - 名前の編集ボタン
 * - 削除ボタン
 */
const PasskeyInfoList: React.FC = () => {
  const { id: userID } = useParams();

  const [userInfo, setUserInfo] = useState<UserInfo>();
  const [PasskeyInfos, setPasskeyInfos] = useState<PasskeyInfo[]>([]);
  // TODO: aaguid.jsonの情報を、フロントエンドでどうやって保持するか要検討
  const [aaguid, setAaguid] = useState<Aaguid>();

  // NOTE: 本当はuseEffectでデータフェッチしたくないけど、ライブラリ入れるのも面倒に感じたので一旦これで…。
  useEffect(() => {
    fetch(`http://localhost:8080/users/${userID}`)
      .then((res) => res.json())
      .then((json) => setUserInfo(json))
      .catch((err) => console.error(err));

    fetch(`http://localhost:8080/users/${userID}/public_keys`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
    })
      .then((res) => res.json())
      .then((json) => setPasskeyInfos(json))
      .catch((err) => console.error(err));

    fetch(
      "https://raw.githubusercontent.com/passkeydeveloper/passkey-authenticator-aaguids/refs/heads/main/aaguid.json"
    )
      .then((res) => res.json())
      .then((json) => setAaguid(json))
      .catch((err) => console.error(err));
  }, [userID]);

  const deletePasskeyInfo = useCallback(
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

      setPasskeyInfos((prevPasskeyInfos) =>
        prevPasskeyInfos.filter((PasskeyInfo) => PasskeyInfo.id !== id)
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
            <th key="AAGUID">AAGUID</th>
            <th>削除する</th>
          </tr>
        </thead>
        <tbody>
          {PasskeyInfos.map((PasskeyInfo) => (
            <tr key={PasskeyInfo.id}>
              <td key="AAGUID">
                <span>
                  {aaguid ? aaguid[PasskeyInfo.AAGUID].name : "Unknown"}
                </span>
              </td>
              <td>
                <button onClick={() => deletePasskeyInfo(PasskeyInfo.id)}>
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

export default PasskeyInfoList;
