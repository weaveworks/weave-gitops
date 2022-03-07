import * as React from "react";
import styled from "styled-components";
import { Divider, Input, InputAdornment, IconButton } from "@material-ui/core";
import { Visibility, VisibilityOff } from "@material-ui/icons";
import Alert from "../components/Alert";
import Button from "../components/Button";
import Flex from "../components/Flex";
import LoadingPage from "../components/LoadingPage";
import { Auth } from "../contexts/AuthContext";
import { theme } from "../lib/theme";
import images from "../lib/images";
import { Redirect } from "react-router-dom";
import { FeatureFlags } from "../contexts/FeatureFlags";

export const SignInPageWrapper = styled(Flex)`
  background: url(${images.SignInBackground});
  height: 100%;
  width: 100%;
`;

export const FormWrapper = styled(Flex)`
  background-color: ${(props) => props.theme.colors.white};
  width: 500px;
  padding-top: ${(props) => props.theme.spacing.medium};
  align-content: space-between;
  .MuiButton-label {
    width: 250px;
  }
  .MuiInputBase-root {
    width: 275px;
  }
  #email,
  #password {
    padding-bottom: ${(props) => props.theme.spacing.xs};
  }
`;

const Logo = styled(Flex)`
  margin-bottom: ${(props) => props.theme.spacing.medium};
`;

const Action = styled(Flex)`
  flex-wrap: wrap;
`;

const Footer = styled(Flex)`
  & img {
    width: 500px;
  }
`;

const AlertWrapper = styled(Alert)`
  .MuiAlert-root {
    width: 470px;
    margin-bottom: ${(props) => props.theme.spacing.small};
  }
`;

function SignIn() {
  const { authFlag } = React.useContext(FeatureFlags);

  if (authFlag === false) {
    return <Redirect to="applications" />;
  }

  const formRef = React.useRef<HTMLFormElement>();
  const { signIn, error, loading, userInfo } = React.useContext(Auth);
  const [password, setPassword] = React.useState<string>("");
  const [showPassword, setShowPassword] = React.useState<boolean>(false);

  const handleOIDCSubmit = () => {
    const CURRENT_URL = window.origin;
    return (window.location.href = `/oauth2?return_url=${encodeURIComponent(
      CURRENT_URL
    )}`);
  };

  const handleUserPassSubmit = () => signIn({ password });

  return (
    <SignInPageWrapper center align column>
      {error && (
        <AlertWrapper
          severity="error"
          title="Error signin in"
          message={`${String(error.status)} ${error.statusText}`}
          center
        />
      )}
      <FormWrapper center align wrap>
        <div style={{ padding: theme.spacing.base }}>
          <Logo>
            <img src={images.WeaveLogo} />
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
            <Flex center align>
              <Input disabled id="email" type="text" value="admin" />
            </Flex>
            <Flex center align>
              <Input
                onChange={(e) => setPassword(e.currentTarget.value)}
                required
                id="password"
                placeholder="Password"
                type={showPassword ? "text" : "password"}
                value={password}
                endAdornment={
                  <InputAdornment position="end">
                    <IconButton
                      aria-label="toggle password visibility"
                      onClick={() => setShowPassword(!showPassword)}
                    >
                      {showPassword ? <Visibility /> : <VisibilityOff />}
                    </IconButton>
                  </InputAdornment>
                }
              />
            </Flex>
            <Flex center>
              {!loading ? (
                <Button
                  type="submit"
                  style={{ marginTop: theme.spacing.medium }}
                >
                  CONTINUE
                </Button>
              ) : (
                <div style={{ margin: theme.spacing.medium }}>
                  <LoadingPage />
                </div>
              )}
            </Flex>
          </form>
        </div>
        <Footer>
          <img src={images.SignInWheel} />
        </Footer>
      </FormWrapper>
    </SignInPageWrapper>
  );
}

export default styled(SignIn)``;
