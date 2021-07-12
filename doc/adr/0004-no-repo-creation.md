# 4. No repo creation

Date: 2021-07-06

## Status

Proposed

## Context

In the initial pre-alpha version of weave-gitops and WKP, we created git repositories for the end-user.  However, we encounter more instances where customers don't permit developers, application or platform SREs to create repositories in the customer's organization.  While this is a convenient feature, we've decided the effort and support required are better focused elsewhere in the application.

## Decision

Weave gitops will require the user to have created the git repository on their Git server before running commands which need us to push changes to that repository.

## Consequences

The automated tests will set up repositories before executing weave-gitops operations against those repositories.
