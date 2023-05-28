import * as React from "react";
import { Auth } from "../../contexts/AuthContext";
import Page from "../../components/Page";
import UserGroupsTable from "../../components/UserGroupsTable";
type Props = {
  className?: string;
};

export default function UserInfo({ className }: Props) {
  const { userInfo, loading } = React.useContext(Auth);
  return (
    <Page
      className={className}
      loading={loading}
      path={[{ label: "User Info" }]}
    >
      <UserGroupsTable rows={userInfo?.groups}></UserGroupsTable>
    </Page>
  );
}
