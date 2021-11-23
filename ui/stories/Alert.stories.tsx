import { Meta, Story } from "@storybook/react";
import * as React from "react";
import Alert from "../components/Alert";

export default {
  title: "Alert",
  parameters: {
    docs: {
      description: {
        component: "MUI Material Alert - see https://mui.com/components/alert/",
      },
    },
  },
  component: Alert,
  argTypes: {
    center: {
      description:
        "Overrides `justify-content: flex-start` (left) to render the Alert in the center of it's 100% width `<Flex />` component",
      defaultValue: false,
      type: { summary: "boolean", name: "boolean", required: false },
    },
    title: {
      description: "text for Mui's `<AlertTitle />` component",
      defaultValue: "Title",
      type: { summary: "string", name: "string", required: false },
    },
    severity: {
      description:
        "string of one of the colors from our `MuiTheme` - also sets the corresponding material icon - see /ui/lib/theme.ts and https://mui.com/customization/theming/",
      defaultValue: "success",
      options: ["success", "error", "warning", "info"],
      type: { summary: "string", name: "string", required: false },
      control: "radio",
    },
    message: {
      description: "Appears under `title`",
      defaultValue: "Message / JSX goes here!",
      type: {
        summary: "string | JSX.Element",
        name: "string",
        required: false,
      },
    },
    className: {
      description: "CSS MUI Overrides or other styling",
      type: { summary: "string", name: "string", required: false },
    },
  },
} as Meta;

const Template: Story = (args) => <Alert {...args} />;

export const Success = Template.bind({});
export const Error = Template.bind({});
Error.args = { severity: "error" };
export const Warning = Template.bind({});
Warning.args = { severity: "warning" };
export const Info = Template.bind({});
Info.args = { severity: "info" };
