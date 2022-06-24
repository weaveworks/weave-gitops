import * as React from "react";
import styled from "styled-components";
import Link from "./Link";

type Props = {
  className?: string;
  image: string;
};

const convertImage = (image: string) => {
  const split = image.split("/");
  const prefix = split.shift();
  let url = "";

  //Github GHCR or Google GCR
  if (prefix === "ghcr.io" || prefix === "gcr.io") return "https://" + image;
  //Quay.io
  else if (prefix === "quay.io") {
    url = "https://quay.io/repository/";
    split.forEach((urlPart, index) => {
      if (index === split.length - 1) url += urlPart;
      else url += `${urlPart}/`;
    });
    return url;
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
      split.forEach((urlPart, index) => {
        if (index === split.length - 1) url += urlPart;
        else url += `${urlPart}/`;
      });
      return url;
    }
  }
  //docker without prefix
  else if (prefix === "library")
    return "https://hub.docker.com/r/_/" + split[1];
  //this one's at risk if we have to add others - global docker images can just be one word apparently
  else if (
    !prefix.includes("public.ecr.aws") &&
    !prefix.includes("amazonaws.com")
  ) {
    if (!split[1]) return "https://hub.docker.com/r/_/" + split[0];
    else {
      url = "https://hub.docker.com/r/";
      split.forEach((urlPart, index) => {
        if (index === split.length - 1) url += urlPart;
        else url += `${urlPart}/`;
      });
      return url;
    }
    //public aws
  } else if (prefix.includes("public.ecr.aws")) {
    url = "https://gallery.ecr.aws/";
    split.forEach((urlPart, index) => {
      if (index === split.length - 1)
        url += urlPart.slice(0, urlPart.indexOf(":"));
      else url += `${urlPart}/`;
    });
    return url;
    //private aws
  } else if (prefix.includes("amazonaws.com")) {
    //OMG WHAT DO WE DO ON THIS ONE
  } else return "";
};

function ImageLink({ className, image = "" }: Props) {
  return <Link className={className}></Link>;
}

export default styled(ImageLink).attrs({ className: ImageLink.name })``;
