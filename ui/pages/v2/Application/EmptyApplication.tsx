import * as React from "react";
import styled from "styled-components";
import FancyCard from "../../../components/FancyCard";
import Flex from "../../../components/Flex";
import Spacer from "../../../components/Spacer";
import images from "../../../lib/images";
import { V2Routes } from "../../../lib/types";
import { addKustomizationURL, formatURL } from "../../../lib/utils";

type Props = {
  className?: string;
  appName: string;
};

function EmptyApplication({ className, appName }: Props) {
  return (
    <Flex className={className} wide center>
      <Spacer m={["none", "none", "medium", "none"]}>
        <FancyCard
          to={addKustomizationURL(appName)}
          image={images.fancyCardBackgroundBlue}
          title="Flux Kustomization"
        >
          We’ll manage your kustomizations for you. Point us to the correct path
          in your repository and we’ll take care of the rest.
        </FancyCard>
      </Spacer>
      <FancyCard
        to={formatURL(V2Routes.AddHelmRelease, { name })}
        image={images.fancyCardBackgroundOrange}
        title="Flux Helm Release"
      >
        Gone are the days of managing helm charts and releases on your own. Add
        a helm release and let Flux do the hard work.
      </FancyCard>
    </Flex>
  );
}

export default styled(EmptyApplication).attrs({
  className: EmptyApplication.name,
})``;
