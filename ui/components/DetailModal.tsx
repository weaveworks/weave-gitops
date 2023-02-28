import React from "react";
import styled from "styled-components";
import { DialogYamlView, YamlViewProps } from "./YamlView";

export enum DetailOptions {
  YamlView = "YamlView",
}

export type PropOptions = YamlViewProps;

export type DetailViewProps = {
  className?: string;
  component: DetailOptions;
  props: PropOptions;
};

function DetailModal({ className, props, component }: DetailViewProps) {
  switch (component) {
    case DetailOptions.YamlView:
      return <DialogYamlView {...props} />;
  }
}

export default styled(DetailModal).attrs({ className: DetailModal.name })``;
