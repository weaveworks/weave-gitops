import * as React from "react";
import { useQuery } from "react-query";
import styled from "styled-components";
// @ts-ignore
import fluxIcon from "../images/flux-icon.svg";

type Props = {
  className?: string;
};

function useDownloadSVG(url: string) {
  return useQuery(
    ["svg", url],
    () => fetch(url).then((response) => response.text()),
    {
      retry: false,
      cacheTime: Infinity,
      staleTime: Infinity,
    }
  );
}

function RemoteSvgIcon({ className }: Props) {
  const { data: svg } = useDownloadSVG(fluxIcon);

  return (
    <div className={className}>
      {svg && <svg dangerouslySetInnerHTML={{ __html: svg }} />}
    </div>
  );
}

export default styled(RemoteSvgIcon).attrs({ className: RemoteSvgIcon.name })``;
