import { Button } from "@material-ui/core";
import _ from "lodash";
import * as React from "react";
import styled from "styled-components";
/*eslint import/no-unresolved: [0]*/
// @ts-ignore
import githubLogo from "url:../images/github.png";
import Page from "../components/Page";
import useAuth from "../hooks/auth";

type Props = {
  className?: string;
};

const LoginContainer = styled.div`
  height: 45px;
  margin-top: 14px;
  text-align: center;
`;

const LoginButton = styled(Button)`
  text-align: left;
  font-weight: bold;
  margin: 0 auto;
  color: ${(props) => props.theme.colors.white};
  display: flex;
  align-items: center;
  padding-left: 3px;
  text-transform: none;
  height: 46px;
  width: 250px;
  font-size: ${(props) => props.theme.fontSizes.normal};
  box-shadow: none;

  &.github {
    background-color: black;
    color: white;
    /* background-image: url(${githubLogo}); */
    &:hover {
      background-color: ${(props) => props.theme.colors.black} !important;
    }
  }

  &.google {
    .fab.fa-google {
      display: none;
    }
    background-repeat: no-repeat;
    padding-left: 65px;
  }

  &.gitlab {
    background-color: orange;
    color: white;
  }
`;

function AuthPage({ className }: Props) {
  const { loading, providers, getProviders, error } = useAuth();

  const doAuth = (url) => {
    window.location.href = url;
  };

  React.useEffect(() => {
    getProviders();
  }, []);

  return (
    <Page error={error} loading={loading} title="Login" className={className}>
      {_.map(providers, (p) => (
        <LoginContainer
          key={p.name}
          onClick={(ev) => {
            ev.preventDefault;
            doAuth(p.authUrl);
          }}
        >
          <LoginButton className={`provider ${p.name}`}>
            Login with {p.name}
          </LoginButton>
        </LoginContainer>
      ))}
    </Page>
  );
}

export default styled(AuthPage)``;
