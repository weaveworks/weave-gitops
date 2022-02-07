import * as React from "react";
import styled from "styled-components";
import Flex from "../components/Flex";
import Page from "../components/Page";
// @ts-ignore
import { ReactComponent as SignInWheel } from "../images/SignInWheel.svg";
// @ts-ignore
import SignInBackground from "./../images/SignInBackground.svg";
import Button from "../components/Button";
import { Auth } from "../contexts/AuthContext";

const PageWrapper = styled(Flex)`
  background: url(${SignInBackground}) no-repeat;
  height: 100vh;
`;

const FormWrapper = styled(Flex)`
  background-color: ${(props) => props.theme.colors.white};
  width: 500px;
  height: 500px;
`;

function SignIn() {
  const { submitAuthType } = React.useContext(Auth);

  const handleSubmit = (event) => {
    event.preventDefault();
    submitAuthType(event.target.data);
  };

  return (
    <PageWrapper center align>
      <FormWrapper center align>
        <Button type="submit" onClick={handleSubmit}>
          LOGIN WITH OIDC PROVIDER
        </Button>
        <Button type="submit" onClick={handleSubmit}>
          LOGIN WITH USERNAME AND PASSWORD
        </Button>
        {/* <SignInWheel /> */}
      </FormWrapper>
    </PageWrapper>
  );
}

export default styled(SignIn)``;
