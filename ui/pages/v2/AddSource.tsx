import * as React from "react";
import styled from "styled-components";
import FancyCard from "../../components/FancyCard";
import Flex from "../../components/Flex";
import Page from "../../components/Page";
import images from "../../lib/images";
import { V2Routes } from "../../lib/types";
import { formatAppScopedURL, formatURL } from "../../lib/utils";

type Props = {
  appName?: string;
  className?: string;
};

function possiblyAppScoped(appName: string, page: V2Routes) {
  return appName ? formatAppScopedURL(appName, page) : formatURL(page);
}

function AddSource({ className, appName }: Props) {
  return (
    <Page
      title={`Add Source${appName ? ` for ${appName}` : ""}`}
      className={className}
    >
      <Flex align wide>
        <FancyCard
          to={possiblyAppScoped(appName, V2Routes.AddGitRepo)}
          image={images.fancyCardBackground}
          title="Git Repository"
        >
          Connect a Git repository from providers such as GitHub and GitLab.
          This gives you the ability to create kustomizations and helm releases.
        </FancyCard>
        <FancyCard
          to={possiblyAppScoped(appName, V2Routes.AddHelmRepo)}
          image={images.fancyCardBackgroundBlue}
          title="Helm Repository"
        >
          Connect to your favorite helm repository to access your favorite helm
          charts via GitOps.
        </FancyCard>
        <FancyCard
          to={possiblyAppScoped(appName, V2Routes.AddBucket)}
          image={images.fancyCardBackgroundOrange}
          title="Bucket"
        >
          Gotta a bunch of yaml files stashed in your favorite cloud provider’s
          bucket system? We can keep a secret!
        </FancyCard>
      </Flex>
    </Page>
  );
}

export default styled(AddSource).attrs({ className: AddSource.name })`
  ${FancyCard} {
    margin-right: ${(props) => props.theme.spacing.large};

    .MuiCardContent-root {
      height: 136px;
      max-width: 272px;
    }
  }
`;
