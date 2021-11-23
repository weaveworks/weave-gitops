import { Meta, Story } from "@storybook/react";
import * as React from "react";
import Alert, { Props } from "../components/Alert";

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
      defaultValue: false,
    },
    severity: {
      defaultValue: "success",
    },
  },
} as Meta;

const Template: Story<Props> = (args) => <Alert {...args} />;

export const Success = Template.bind({});
Success.args = { title: "Title", message: "Message / JSX" };
export const Error = Template.bind({});
Error.args = { ...Success.args, severity: "error" };
export const Warning = Template.bind({});
Warning.args = { ...Success.args, severity: "warning" };
export const Info = Template.bind({});
Info.args = { ...Success.args, severity: "info" };
