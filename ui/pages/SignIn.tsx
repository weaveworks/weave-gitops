import { Visibility, VisibilityOff } from "@mui/icons-material";
import { IconButton, Input, InputAdornment } from "@mui/material";
import qs from "query-string";
import * as React from "react";
import styled from "styled-components";
import Alert from "../components/Alert";
import Button from "../components/Button";
import DarkModeSwitch from "../components/DarkModeSwitch";
import Flex from "../components/Flex";
import LoadingPage from "../components/LoadingPage";
import { Auth, AuthContext } from "../contexts/AuthContext";
import { useFeatureFlags } from "../hooks/featureflags";
import { useInDarkMode } from "../hooks/theme";
import images from "../lib/images";
import { withBasePath } from "../lib/utils";

export const FormWrapper = styled(Flex)`
  background-color: ${(props) => props.theme.colors.neutral00};
  border-radius: ${(props) => props.theme.borderRadius.soft};
  width: 500px;
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
  margin-bottom: ${(props) => props.theme.spacing.large};
  img:first-child {
    margin-right: ${(props) => props.theme.spacing.xs};
  }
`;

const OidcFlex = styled(Flex)`
  ${(props) =>
    props.clusterAuth && `padding-bottom: ${props.theme.spacing.small}`}
`;

const Footer = styled(Flex)`
  & img {
    width: 500px;
    height: 300px;
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

const SwitchFlex = styled(Flex)`
  padding: ${(props) => props.theme.spacing.small};
  box-sizing: border-box;
`;

const MarginButton = styled(Button)`
  &.MuiButtonBase-root {
    margin-top: ${(props) => props.theme.spacing.medium};
  }
`;
type Props = {
  darkModeEnabled?: boolean;
};
function SignIn({ darkModeEnabled = true }: Props) {
  const { isFlagEnabled, flags } = useFeatureFlags();

  const formRef = React.useRef<HTMLFormElement>(undefined);
  const {
    signIn,
    error: authError,
    loading: authLoading,
  } = React.useContext(Auth) as AuthContext;
  const [password, setPassword] = React.useState<string>("");
  const [username, setUsername] = React.useState<string>("");
  const [showPassword, setShowPassword] = React.useState<boolean>(false);

  const handleOIDCSubmit = () => {
    const redirect = qs.parse(window.location.search).redirect || "";

    // Head to the BE to start the OIDC flow so we do not use any of
    // react-router or other client-side routing
    return (window.location.href =
      withBasePath("/oauth2?") +
      qs.stringify({
        // BE handles the redirect to return_url after authentication
        // so add the base path
        return_url: window.origin + withBasePath(redirect.toString()),
      }));
  };

  const handleUserPassSubmit = () => signIn({ username, password });

  const dark = useInDarkMode();

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
      {authError && (
        <AlertWrapper
          severity="error"
          title="Error signing in"
          message={`${
            authError.status === 401
              ? `Incorrect username or password.`
              : `${authError.status} ${authError.statusText}`
          }`}
          center
        />
      )}
      <FormWrapper
        column
        align
        wrap
        style={{
          zIndex: 999,
        }}
      >
        <SwitchFlex wide align end>
          <DarkModeSwitch darkModeEnabled={darkModeEnabled} />
        </SwitchFlex>
        <Logo wide align center>
          <img
            src={dark ? images.logoDark : images.logoLight}
            height="60px"
            width="auto"
            alt=""
          />
          <img
            src={dark ? images.logotypeLight : images.logotype}
            height="32px"
            width="auto"
            alt=""
          />
        </Logo>
        {isFlagEnabled("OIDC_AUTH") ? (
          <OidcFlex
            wide
            center
            //extra padding-bottom for when both auth flags are enabled
            clusterAuth={isFlagEnabled("CLUSTER_USER_AUTH")}
          >
            <Button
              type="submit"
              onClick={(e) => {
                e.preventDefault();
                handleOIDCSubmit();
              }}
            >
              {flags.WEAVE_GITOPS_FEATURE_OIDC_BUTTON_LABEL ||
                "LOGIN WITH OIDC PROVIDER"}
            </Button>
          </OidcFlex>
        ) : null}
        {isFlagEnabled("CLUSTER_USER_AUTH") ? (
          <form
            ref={formRef}
            onSubmit={(e) => {
              e.preventDefault();
              handleUserPassSubmit();
            }}
          >
            <Flex wide tall column align>
              <Input
                onChange={(e) => setUsername(e.currentTarget.value)}
                id="email"
                type="text"
                placeholder="Username"
                value={username}
                required
                autoFocus
              />
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
                      color="primary"
                      size="large"
                    >
                      {showPassword ? <Visibility /> : <VisibilityOff />}
                    </IconButton>
                  </InputAdornment>
                }
              />
              {!authLoading ? (
                <MarginButton type="submit">CONTINUE</MarginButton>
              ) : (
                <LoadingPage />
              )}
            </Flex>
          </form>
        ) : null}
        <DocsWrapper center align>
          Need help? Have a look at the&nbsp;
          <a
            href="https://docs.gitops.weaveworks.org/docs/getting-started"
            target="_blank"
            rel="noopener noreferrer"
          >
            documentation.
          </a>
        </DocsWrapper>
        <Footer>
          <img
            src={dark ? images.signInWheelDark : images.signInWheel}
            alt=""
          />
        </Footer>
      </FormWrapper>
    </Flex>
  );
}

export default styled(SignIn)`
  ${LoadingPage} {
    padding: ${(props) => props.theme.spacing.medium};
  }
`;
