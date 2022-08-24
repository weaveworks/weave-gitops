import { Divider, IconButton, Input, InputAdornment } from "@material-ui/core";
import Visibility from "@material-ui/icons/Visibility";
import VisibilityOff from "@material-ui/icons/VisibilityOff";
import * as React from "react";
import styled from "styled-components";
import Alert from "../components/Alert";
import Button from "../components/Button";
import Flex from "../components/Flex";
import LoadingPage from "../components/LoadingPage";
import { Auth } from "../contexts/AuthContext";
import { useFeatureFlags } from "../hooks/featureflags";
import images from "../lib/images";
import { theme } from "../lib/theme";

const SignInBackgroundAnimation = React.lazy(
  () => import("../components/Animations/SignInBackground")
);

export const FormWrapper = styled(Flex)`
  background-color: ${(props) => props.theme.colors.white};
  border-radius: ${(props) => props.theme.borderRadius.soft};
  width: 500px;
  padding-top: ${(props) => props.theme.spacing.large};
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

const Footer = styled(Flex)`
  & img {
    width: 500px;
  }
`;

const AlertWrapper = styled(Alert)`
  width: auto;
  .MuiAlert-root {
    width: 470px;
    margin-bottom: ${(props) => props.theme.spacing.small};
  }
`;

const DocsWrapper = styled(Flex)`
  padding: ${(props) => props.theme.spacing.medium};
  font-size: ${({ theme }) => theme.fontSizes.small};
  a {
    color: ${({ theme }) => theme.colors.primary};
  }
`;

function SignIn() {
  const { data } = useFeatureFlags();
  const flags = data?.flags || {};

  const formRef = React.useRef<HTMLFormElement>();
  const {
    signIn,
    error: authError,
    loading: authLoading,
  } = React.useContext(Auth);
  const [password, setPassword] = React.useState<string>("");
  const [username, setUsername] = React.useState<string>("");
  const [showPassword, setShowPassword] = React.useState<boolean>(false);

  const handleOIDCSubmit = () => {
    const CURRENT_URL = window.origin;
    return (window.location.href = `/oauth2?return_url=${encodeURIComponent(
      CURRENT_URL
    )}`);
  };

  const handleUserPassSubmit = () => signIn({ username, password });

  return (
    <Flex
      tall
      wide
      center
      align
      column
      style={{
        height: "100vh",
        width: "100vw",
      }}
    >
      <React.Suspense fallback={null}>
        <SignInBackgroundAnimation />
      </React.Suspense>
      {authError && (
        <AlertWrapper
          severity="error"
          title="Error signin in"
          message={`${
            authError.status === 401
              ? `Incorrect username or password.`
              : `${authError.status} ${authError.statusText}`
          }`}
          center
        />
      )}
      <FormWrapper
        center
        align
        wrap
        style={{
          zIndex: 999,
        }}
      >
        <div>
          <Logo wide center>
            <img src={images.weaveLogo} />
          </Logo>
          {flags.OIDC_AUTH ? (
            <Flex wide center>
              <Button
                type="submit"
                onClick={(e) => {
                  e.preventDefault();
                  handleOIDCSubmit();
                }}
              >
                LOGIN WITH OIDC PROVIDER
              </Button>
            </Flex>
          ) : null}
          {flags.OIDC_AUTH && flags.CLUSTER_USER_AUTH ? (
            <Divider variant="middle" style={{ margin: theme.spacing.base }} />
          ) : null}
          {flags.CLUSTER_USER_AUTH ? (
            <form
              ref={formRef}
              onSubmit={(e) => {
                e.preventDefault();
                handleUserPassSubmit();
              }}
            >
              <Flex center align>
                <Input
                  onChange={(e) => setUsername(e.currentTarget.value)}
                  id="email"
                  type="text"
                  placeholder="Username"
                  value={username}
                  required
                />
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
                {!authLoading ? (
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
          ) : null}
          <DocsWrapper center align>
            Need help? Have a look at the&nbsp;
            <a
              href="https://docs.gitops.weave.works/docs/getting-started"
              target="_blank"
              rel="noopener noreferrer"
            >
              documentation.
            </a>
          </DocsWrapper>
        </div>
        <Footer>
          <img src={images.signInWheel} />
        </Footer>
      </FormWrapper>
    </Flex>
  );
}

export default styled(SignIn)``;
