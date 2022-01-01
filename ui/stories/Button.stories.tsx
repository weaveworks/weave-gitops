import { Meta, Story } from "@storybook/react";
import React from "react";
import Button, { Props } from "../components/Button";

export default {
  title: "Button",
  component: Button,
  parameters: {
    docs: {
      description: {
        component:
          "MUI Material Button - see https://mui.com/components/buttons/ for full props.",
      },
    },
  },
  argTypes: {
    color: { table: { defaultValue: { summary: "primary" } } },
    variant: { table: { defaultValue: { summary: "outlined" } } },
    loading: {
      defaultValue: false,
      table: { defaultValue: { summary: "false" } },
    },
    startIcon: {
      description: "`<Icon />` Element to come before `<Button />` content",
      table: {
        type: { summary: "<Icon type={IconType.YOUR_TYPE} size='base' />" },
      },
      control: null,
    },
    onClick: {
      description: "Event Handler",
      type: { summary: "function(e) {}", name: "function" },
      action: "click",
    },
    disableElevation: {
      defaultValue: true,
      table: { defaultValue: { summary: "true" } },
    },
    disabled: {
      defaultValue: false,
      table: { defaultValue: { summary: "false" } },
    },
    className: {},
  },
} as Meta;

const Template: Story<Props> = (args) => <Button {...args}>Weaveworks</Button>;

export const Default = Template.bind({});
Default.args = { color: "primary", variant: "outlined" };
export const ModalButton = Template.bind({});
ModalButton.args = {
  color: "inherit",
  variant: "text",
};
