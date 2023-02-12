import * as React from "react";
import { Auth } from "../../contexts/AuthContext";

import Page from "../../components/Page";
import UserGroupsTable from "../../components/UserGroupsTable";
type Props = {
  className?: string;
};

export default function ProfileInfo({ className }: Props) {
  const { userInfo, error, loading } = React.useContext(Auth);
  console.log({ userInfo });
  return (
    <Page className={className} loading={loading}>
      <UserGroupsTable
        rows={[
          "asd@gmail.com",
          "HelloWorld@gmail.com",
          "Ahmed Magdy@gmail.com",
          "Ali@gmail.com",
          "Simon@gmail.com",
          "Alina@gmail.com",
          "David@gmail.com",
          "HelloWorld@gmail.com",
          "Ahmed Magdy@gmail.com",
          "Ali@gmail.com",
          "Simon@gmail.com",
          "Alina@gmail.com",
        ]}
      ></UserGroupsTable>
    </Page>
  );
}
