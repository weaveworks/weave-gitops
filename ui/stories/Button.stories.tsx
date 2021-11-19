import { CircularProgress } from "@material-ui/core";
import React from "react";
import { theme } from "..";
import Button from "../components/Button";
import Flex from "../components/Flex";

export default {
  title: "Button",
  component: Button,
};

const Template = (args) => <Button {...args}>{args.content}</Button>;

export const Primary = Template.bind({});
export const Loading = Template.bind({});
export const Secondary = Template.bind({});

Primary.args = {
  variant: "outlined",
  color: "primary",
};

Secondary.args = {
  content: "Storybook",
  variant: "outlined",
  color: "secondary",
};

Loading.args = {
  content: (
    <Flex wide align>
      <CircularProgress size={theme.fontSizes.normal} />
    </Flex>
  ),
  disabled: true,
  variant: "outlined",
  color: "primary",
};
