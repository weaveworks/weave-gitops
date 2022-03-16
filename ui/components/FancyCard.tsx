import { Card, CardContent, CardMedia } from "@material-ui/core";
import * as React from "react";
import styled from "styled-components";
import images from "../lib/images";
import Flex from "./Flex";
import Link from "./Link";
import Text from "./Text";

type Props = {
  to: string;
  className?: string;
  children?: any;
  title: string;
  image: keyof typeof images;
};

const ImageContainer = styled.div`
  position: relative;

  /* img {
    width: 272px;
    height: 137px;
  } */
`;

const TitleContainer = styled(Flex)`
  position: absolute;
  height: 100%;
  color: white;
  text-align: center;
`;

function FancyCard({ className, to, children, title, image }: Props) {
  return (
    <div className={className}>
      <Link to={to}>
        <Card elevation={2}>
          <ImageContainer>
            <TitleContainer wide align center>
              <div>
                <Text bold size="large">
                  {title}
                </Text>
              </div>
            </TitleContainer>

            <CardMedia component="img" image={image} />
          </ImageContainer>
          <CardContent>{children}</CardContent>
        </Card>
      </Link>
    </div>
  );
}

export default styled(FancyCard).attrs({ className: FancyCard.name })`
  .MuiCardContent-root {
    box-sizing: border-box;
  }
`;
