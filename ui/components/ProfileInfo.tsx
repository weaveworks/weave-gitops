import * as React from "react";
import styled from "styled-components";
import { Auth } from "../contexts/AuthContext";

import Page from "./Page";
type Props = {
  className?: string;
};

export default function Notifications({ className }: Props) {
  const { userInfo, error, loading } = React.useContext(Auth);
  console.log({userInfo})
  return (
    <Page
      className={className}
      loading={loading}
    >
  
    </Page>
  );
}
