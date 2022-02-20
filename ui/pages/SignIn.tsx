import * as React from "react";
import styled from "styled-components";
import { theme } from "../lib/theme";
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
    height: ${(props) => props.theme.spacing.xl};
  }
`;

function SignIn() {
  const formRef = React.useRef<HTMLFormElement>();
  const { signIn } = React.useContext(Auth);
  const [password, setPassword] = React.useState<string>("");

  const handleOIDCSubmit = () => {
    const CURRENT_URL = window.origin;
    console.log(CURRENT_URL);
    return (window.location.href = `/oauth2?return_url=${encodeURIComponent(
      CURRENT_URL
    )}`);
  };

  const handleUserPassSubmit = () => signIn({ password });

  return (
    <PageWrapper center align>
      <FormWrapper center align wrap>
        <div style={{ padding: theme.spacing.base }}>
          <Logo>
            <img src={WeaveLogo} />
          </Logo>
          <Action>
            <Button
              type="submit"
              onClick={(e) => {
                e.preventDefault();
                handleOIDCSubmit();
              }}
            >
              LOGIN WITH OIDC PROVIDER
            </Button>
            <Divider
              variant="middle"
              style={{ width: "100%", margin: theme.spacing.base }}
            />
          </Action>
          <form
            ref={formRef}
            onSubmit={(e) => {
              e.preventDefault();
              handleUserPassSubmit();
            }}
          >
            <FormElement center align>
              <TextField
                onChange={(e) => setPassword(e.currentTarget.value)}
                required
                id="password"
                label="Password"
                variant="standard"
                // type="password"
                value={password}
              />
            </FormElement>
            <Flex center>
              <Button type="submit" style={{ marginTop: theme.spacing.medium }}>
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
