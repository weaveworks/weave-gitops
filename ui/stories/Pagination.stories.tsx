import { Meta, Story } from "@storybook/react";
import React from "react";
import Pagination, { Props } from "../components/Pagination";

export default {
  title: "Pagination",
  component: Pagination,
  parameters: {
    docs: {
      description: {
        component: "Pagination controller for DataTable",
      },
    },
  },
} as Meta;

const Template: Story<Props> = (args) => <Pagination {...args} />;

export const Default = Template.bind({});
Default.args = {
  className: "",
  onForward: () => "",
  onSkipForward: () => "",
  onBack: () => "",
  onSkipBack: () => "",
  onSelect: () => "",
  index: 0,
  length: 1,
  totalObjects: 1,
  perPageOptions: [25, 50, 75, 100],
};
