import * as React from "react";
import styled from "styled-components";
import Flex from "../components/Flex";
// @ts-ignore
import SignInWheel from "url:../images/SignInWheel.svg";
// @ts-ignore
import SignInBackground from "./../images/SignInBackground.svg";
// @ts-ignore
import WeaveLogo from "./../images/WeaveLogo.svg";
import Button from "../components/Button";
import { Auth } from "../contexts/AuthContext";

const PageWrapper = styled(Flex)`
  background: url(${SignInBackground});
  height: 100%;
  width: 100%;
`;

const FormWrapper = styled(Flex)`
  background-color: ${(props) => props.theme.colors.white};
  width: 500px;
  min-height: 70%;
  padding-top: ${(props) => props.theme.spacing.xl};
  align-content: space-between;
  border-radius: ${(props) => props.theme.borderRadius.soft};
`;

const Logo = styled(Flex)`
  margin-bottom: ${(props) => props.theme.spacing.medium};
`;

const Action = styled(Flex)`
  flex-wrap: wrap;
  & button {
    width: 300px;
    margin: ${(props) => props.theme.spacing.xs};
  }
`;

const Footer = styled(Flex)`
  & img {
    max-width: 500px;
  }
`;

function SignIn() {
  const { submitAuthType } = React.useContext(Auth);

  const handleSubmit = (event) => {
    event.preventDefault();
    submitAuthType(event.target.data);
  };

  return (
    <PageWrapper center align>
      <FormWrapper center align wrap>
        <div style={{ marginBottom: "48px" }}>
          <Logo>
            <img src={WeaveLogo} />
          </Logo>
          <Action>
            <Button type="submit" onClick={handleSubmit}>
              LOGIN WITH OIDC PROVIDER
            </Button>
            <Button type="submit" onClick={handleSubmit}>
              LOGIN WITH USERNAME AND PASSWORD
            </Button>
          </Action>
        </div>
        <Footer>
          <img src={SignInWheel} />
        </Footer>
      </FormWrapper>
    </PageWrapper>
  );
}

export default styled(SignIn)``;
