import * as React from "react";
import styled from "styled-components";
import { convertImage } from "../lib/utils";
import Link from "./Link";
import Text from "./Text";

type Props = {
  className?: string;
  image: string;
};

function ImageLink({ className, image = "" }: Props) {
  const imageUrl = convertImage(image);
  if (imageUrl && image !== "-")
    return (
      <Link className={className} href={imageUrl} newTab>
        {image}
      </Link>
    );
  else return <Text>{image}</Text>;
}

export default styled(ImageLink).attrs({ className: ImageLink.name })``;
