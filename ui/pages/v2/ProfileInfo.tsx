import * as React from "react";
import { Auth } from "../../contexts/AuthContext";
import Page from "../../components/Page";
import UserGroupsTable from "../../components/UserGroupsTable";
type Props = {
  className?: string;
};

export default function ProfileInfo({ className }: Props) {
  const { userInfo, loading } = React.useContext(Auth);
  return (
    <Page className={className} loading={loading}>
      <UserGroupsTable rows={userInfo?.groups}></UserGroupsTable>
    </Page>
  );
}
