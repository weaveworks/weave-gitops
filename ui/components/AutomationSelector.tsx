import * as React from "react";
import styled from "styled-components";
import images from "../lib/images";
import { formatURL } from "../lib/nav";
import { V2Routes } from "../lib/types";
import FancyCard from "./FancyCard";
import Flex from "./Flex";
import Spacer from "./Spacer";

type Props = {
  className?: string;
  appName?: string;
};

function AutomationSelector({ className, appName }: Props) {
  return (
    <Flex className={className} wide center>
      <Spacer m={["none", "none", "medium", "none"]}>
        <FancyCard
          to={formatURL(V2Routes.NotImplemented, { appName })}
          image={images.fancyCardBackgroundBlue}
          title="Flux Kustomization"
        >
          We’ll manage your kustomizations for you. Point us to the correct path
          in your repository and we’ll take care of the rest.
        </FancyCard>
      </Spacer>
      <FancyCard
        to={formatURL(V2Routes.NotImplemented, { appName })}
        image={images.fancyCardBackgroundOrange}
        title="Flux Helm Release"
      >
        Gone are the days of managing helm charts and releases on your own. Add
        a helm release and let Flux do the hard work.
      </FancyCard>
    </Flex>
  );
}

export default styled(AutomationSelector).attrs({
  className: AutomationSelector.name,
})`
  ${FancyCard} {
    max-width: 272px;
  }
`;
