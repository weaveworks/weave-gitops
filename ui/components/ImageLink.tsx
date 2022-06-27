import * as React from "react";
import styled from "styled-components";
import Link from "./Link";
import Text from "./Text";

export const convertImage = (image: string) => {
  const split = image.split("/");
  const prefix = split.shift();
  let url = "";

  const makeUrl = (url: string, parts: string[]) => {
    parts.forEach((part, index) => {
      if (index === split.length - 1) url += part;
      else url += `${part}/`;
    });
    return url;
  };

  //Github GHCR or Google GCR
  if (prefix === "ghcr.io" || prefix === "gcr.io") return "https://" + image;
  //Quay.io
  else if (prefix === "quay.io") {
    url = "https://quay.io/repository/";
    return makeUrl(url, split);
  }
  //complex docker prefix case
  else if (prefix === "docker.io") {
    url = "https://hub.docker.com/r/";
    //library alias
    if (split[0] === "library") return url + "_/" + split[1];
    //global
    else if (!split[1]) return url + "_/" + split[0];
    //namespaced
    else {
      return makeUrl(url, split);
    }
  }
  //docker without prefix
  else if (prefix === "library")
    return "https://hub.docker.com/r/_/" + split[0];
  //this one's at risk if we have to add others - global docker images can just be one word apparently
  else if (
    !prefix.includes("public.ecr.aws") &&
    !prefix.includes("amazonaws.com")
  ) {
    if (!split[0]) return "https://hub.docker.com/r/_/" + image;
    else {
      return "https://hub.docker.com/r/" + image;
    }
  } else return "";
};

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
