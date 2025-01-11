import { TextEncoder, TextDecoder } from 'util';
import failOnConsole from 'jest-fail-on-console'

global.TextEncoder = TextEncoder;
global.TextDecoder = TextDecoder as typeof global.TextDecoder;

failOnConsole()
