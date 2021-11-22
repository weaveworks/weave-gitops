import { Meta, Story } from "@storybook/react";
import * as React from "react";
import Alert from "../components/Alert";

export default {
  title: "Alert",
  component: Alert,
  argTypes: {
    center: {
      description:
        "Overrides `justify-content: flex-start` (left) to render the Alert in the center of it's 100% width `<Flex />` component",
      defaultValue: false,
      type: { summary: "boolean", name: "boolean", required: false },
    },
    title: {
      description: "text for Mui's `<AlertTitle />` component - see",
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
      description: "Modal Content",
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

export const Default = Template.bind({});
