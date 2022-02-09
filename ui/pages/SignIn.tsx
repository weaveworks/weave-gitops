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
import { TextField, Divider } from "@material-ui/core";

const PageWrapper = styled(Flex)`
  background: url(${SignInBackground});
  height: 100%;
  width: 100%;
`;

const FormWrapper = styled(Flex)`
  background-color: ${(props) => props.theme.colors.white};
  width: 500px;
  padding-top: ${(props) => props.theme.spacing.medium};
  align-content: space-between;
  border-radius: ${(props) => props.theme.borderRadius.soft};
  & button {
    width: 300px;
    margin: ${(props) => props.theme.spacing.xs};
  }
`;

const Logo = styled(Flex)`
  margin-bottom: ${(props) => props.theme.spacing.small};
`;

const Action = styled(Flex)`
  flex-wrap: wrap;
`;

const Footer = styled(Flex)`
  & img {
    width: 500px;
  }
`;

const FormElement = styled(Flex)`
  .MuiFormControl-root {
    min-width: 300px;
    height: 48px;
  }
`;

function SignIn() {
  const formRef = React.useRef<HTMLFormElement>();
  const { signIn } = React.useContext(Auth);
  const [formState, setFormState] = React.useState<{
    username: string;
    password: string;
  }>({ username: "", password: "" });

  const handleSubmit = (type: string) => {
    console.log(type, { ...formState });
    signIn(type, formState.username, formState.password);
  };

  return (
    <PageWrapper center align>
      <FormWrapper center align wrap>
        <div style={{ padding: "16px" }}>
          <Logo>
            <img src={WeaveLogo} />
          </Logo>
          <Action>
            <Button
              type="submit"
              onClick={(e) => {
                e.preventDefault();
                handleSubmit("oidc");
              }}
            >
              LOGIN WITH OIDC PROVIDER
            </Button>
            <Divider
              variant="middle"
              style={{ width: "100%", margin: "16px" }}
            />
          </Action>

          <form
            ref={formRef}
            onSubmit={(e) => {
              e.preventDefault();
              handleSubmit("username");
            }}
          >
            <FormElement center align>
              <TextField
                onChange={(e) => {
                  setFormState({
                    ...formState,
                    username: e.currentTarget.value,
                  });
                }}
                required
                id="username"
                label="Username"
                variant="standard"
                value={formState.username}
              />
            </FormElement>
            <FormElement center align>
              <TextField
                onChange={(e) => {
                  setFormState({
                    ...formState,
                    password: e.currentTarget.value,
                  });
                }}
                required
                id="password"
                label="Password"
                variant="standard"
                value={formState.password}
              />
            </FormElement>
            <Flex center>
              <Button type="submit" style={{ marginTop: "24px" }}>
                CONTINUE
              </Button>
            </Flex>
          </form>
        </div>
        <Footer>
          <img src={SignInWheel} />
        </Footer>
      </FormWrapper>
    </PageWrapper>
  );
}

export default styled(SignIn)``;
