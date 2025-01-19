import eslint from '@eslint/js';
import tseslint from 'typescript-eslint';
import importPlugin from 'eslint-plugin-import';
import reactPlugin from 'eslint-plugin-react';
import reactHooksPlugin from 'eslint-plugin-react-hooks';

export default tseslint.config([
    {
        ignores: [
            "**/*.pb.ts",
            "**/*.js",
            "**/*.mjs"
        ],
    },
    {
        extends: [
            eslint.configs.recommended,
            importPlugin.flatConfigs.errors,
            importPlugin.flatConfigs.warnings,
            tseslint.configs.recommended,
            importPlugin.flatConfigs.typescript,
            reactPlugin.configs.flat['jsx-runtime'],
            reactHooksPlugin.configs['recommended-latest'],
        ],
        ...reactPlugin.configs.flat.recommended,
    },
    {
        rules: {
            "import/named": 2,
            "import/order": [2,
                {
                    alphabetize: {
                        order: "asc",
                        caseInsensitive: true,
                    },
                }
            ],
            "@typescript-eslint/no-explicit-any": 0,
            "import/no-named-as-default": 0,
            "@typescript-eslint/switch-exhaustiveness-check": [2,
                {
                    "considerDefaultExhaustiveForUnions": true
                }],
            "import/no-unresolved": 0,

            "react/display-name": 0,
            "react/prop-types": 0,
            "react-hooks/exhaustive-deps": 0,
            "react-hooks/rules-of-hooks": 0,
        },
        settings: {
            react: {
                version: "detect",
            },
        },
        languageOptions: {
            ecmaVersion: 2018,
            sourceType: "commonjs",
            parserOptions: {
                project: "./tsconfig.json",
            },
        },
    }
]);
