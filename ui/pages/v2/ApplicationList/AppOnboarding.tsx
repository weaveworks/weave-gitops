import { Card, CardContent, CardMedia } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import Flex from "../../../components/Flex";
import Link from "../../../components/Link";
import Text from "../../../components/Text";
import images from "../../../lib/images";
import { V2Routes } from "../../../lib/types";

type Props = {
  className?: string;
};

const ImageContainer = styled.div`
  position: relative;
`;

const TextContainer = styled(Flex)`
  position: absolute;
  height: 100%;
  color: white;
  text-align: center;
`;

function AppOnboarding({ className }: Props) {
  return (
    <div className={className}>
      <Flex wide center>
        <Link to={V2Routes.NewApp}>
          <Card elevation={2}>
            <ImageContainer>
              <TextContainer wide align center>
                <div>
                  <Text bold size="large">
                    New Application
                  </Text>
                </div>
              </TextContainer>

              <CardMedia component="img" image={images.fancyCardBackground} />
            </ImageContainer>
            <CardContent>
              Get started with your GitOps journey. We’ll walk you through
              setting up your first application. An application is a collection
              of components that you can easily create, manage, and deploy via
              GitOps.
            </CardContent>
          </Card>
        </Link>
      </Flex>
    </div>
  );
}

export default styled(AppOnboarding).attrs({ className: AppOnboarding.name })`
  .MuiCard-root {
    cursor: pointer;
    width: 272px;
  }
`;
