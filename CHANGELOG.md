# [](https://github.com/symphony09/running/compare/v0.3.1...v) (2023-08-06)


### Bug Fixes

* fix missing data during serialization ([ac6a99d](https://github.com/symphony09/running/commit/ac6a99db9e475acf44e336a3ce85cc7986e6b2a3))


### Features

* auto set builder info by util ([1d0d88f](https://github.com/symphony09/running/commit/1d0d88f84f1c6957b09d0ef8624e744f887fa430))
* support set and inspect node builder info ([59ed020](https://github.com/symphony09/running/commit/59ed02095f1164565d57a6c80df9ca0918ae0236))
* update engine inspector ([5efbb8f](https://github.com/symphony09/running/commit/5efbb8f05a1730751866010ef9d89bf24ad7c3b5))



## [0.3.1](https://github.com/symphony09/running/compare/v0.2.5...v0.3.1) (2023-03-26)


### Features

* add hook state impl ([ef5800d](https://github.com/symphony09/running/commit/ef5800d2e70d1e6161dab77da1063e2326ef6f83))
* support add virtual nodes ([181f267](https://github.com/symphony09/running/commit/181f26744ff2a8370f75b510cc87c114d82eb61b))
* support match nodes to run by labels ([f5a703d](https://github.com/symphony09/running/commit/f5a703d584fe044939b9beee6f7ca5580055a22b))
* support passing state in the context ([2f86481](https://github.com/symphony09/running/commit/2f86481c4280508e1dff7a571a1f671a074e2b9d))
* support skip nodes by set ctx params ([f64578a](https://github.com/symphony09/running/commit/f64578a23e8532b02d6c1f400cccf13f492fc105))
* support skip on ctx err ([4c86f00](https://github.com/symphony09/running/commit/4c86f00c9adc6d2060b57c6bd615de3aa8719958))



## [0.2.5](https://github.com/symphony09/running/compare/v0.2.4...v0.2.5) (2023-01-31)


### Bug Fixes

* clear the prev count before recount ([342d7cd](https://github.com/symphony09/running/commit/342d7cd674dcfbb493ff549ca89acbe85b388156))



## [0.2.4](https://github.com/symphony09/running/compare/v0.2.3...v0.2.4) (2023-01-30)


### Bug Fixes

* fix bugs in debug wrapper and base test ([e6e2393](https://github.com/symphony09/running/commit/e6e2393054a5b45f709a4e0279436d7c074d26d3))


### Performance Improvements

* add node name to panic info ([7bd6475](https://github.com/symphony09/running/commit/7bd6475d96595dfa2216807314c70c8c2eabacf0))
* debug wrapper cover more panic case ([2b5ce7b](https://github.com/symphony09/running/commit/2b5ce7b7e70ed207576209e7fe383306b27110d7))
* handle build worker panic ([1c614e0](https://github.com/symphony09/running/commit/1c614e01ec522bf646fce312523885c38548b567))
* perf error handing ([9780101](https://github.com/symphony09/running/commit/978010115d37c17cbada42695c3fb1937db0ab2f))
* perf node register utils ([319c237](https://github.com/symphony09/running/commit/319c237d4803190ad922f76fb761d14508d54b81))



## [0.2.3](https://github.com/symphony09/running/compare/v0.2.2...v0.2.3) (2022-11-11)


### Features

* support inspect engine info ([5602af4](https://github.com/symphony09/running/commit/5602af40f185d53ca18c5e69cd0aab8eaa8cc573))
* support register node directly by utils ([6ba34c7](https://github.com/symphony09/running/commit/6ba34c77a04f6456a8db6bc937377494c41c61db))
* support sub props by utils ([9ae8a3b](https://github.com/symphony09/running/commit/9ae8a3b9588aecb9ed6dae860bc9a4db75d2c0b2))


### Performance Improvements

* perf node register utils ([07b6bed](https://github.com/symphony09/running/commit/07b6bed31accd1cf6c71f35690c72ad130ebdaf6))
* perf props serialization logic ([dc03833](https://github.com/symphony09/running/commit/dc0383300d67219f942ce76cfa7606a6e5ee5b1e))



## [0.2.2](https://github.com/symphony09/running/compare/v0.2.1...v0.2.2) (2022-10-23)


### Bug Fixes

* fix register node builder dead lock ([c5af743](https://github.com/symphony09/running/commit/c5af7437c5e0b0a9018a6e8fe8104b81057d560f))
* fix the logic of prebuilt and reuse node ([997d9df](https://github.com/symphony09/running/commit/997d9dfca7709638a904032ade74e6387a020116))


### Features

* add wrap all nodes option ([3619b2c](https://github.com/symphony09/running/commit/3619b2c3fe7a688136210ab4c62abeb355b682d6))
* support warmup worker pool ([f86467b](https://github.com/symphony09/running/commit/f86467bf2cbd0abc17b45f574141c2743243ba7d))


### Performance Improvements

* perf log content of debug wrapper ([03ba830](https://github.com/symphony09/running/commit/03ba8309ca00a4b67fec638ffc32efc9eef1290f))



## [0.2.1](https://github.com/symphony09/running/compare/v0.1.2...v0.2.1) (2022-10-15)


### Bug Fixes

* fix data race ([078a38d](https://github.com/symphony09/running/commit/078a38d8e7775021c91ed846b4fad1b9611bc258))


### Features

* add async wrapper ([f946f79](https://github.com/symphony09/running/commit/f946f7998437aa6d492e5a94b1e21a6a0e4883c4))
* add debug wrapper ([b68be14](https://github.com/symphony09/running/commit/b68be145c824d45db3f956f379ee41bad98a2474))
* add transactional cluster ([4e1ec5a](https://github.com/symphony09/running/commit/4e1ec5ac2132d6131ea32f2c187062ca69bf402d))
* implement reversible for base node ([aa7ac04](https://github.com/symphony09/running/commit/aa7ac043be3e1f3cf771fba8af2373136f719846))
* support export plan from engine ([e9656ac](https://github.com/symphony09/running/commit/e9656ac85447e7898b5c2d29fd0d9f8d41a474c5))


### Performance Improvements

* perf panic handler in async wrapper ([7508db8](https://github.com/symphony09/running/commit/7508db8b7cac6064ad96c3fcd6c4f48f8b2961dd))
* perf props serialization ([482b6bf](https://github.com/symphony09/running/commit/482b6bf15ce92f05ab2c16ccc5f7d61e5fcc4c60))
* perf set worker pool ([8b48f42](https://github.com/symphony09/running/commit/8b48f42e6cf0f1e1d772aac203f167cf86f80eba))
* perf simple node builder ([e510f71](https://github.com/symphony09/running/commit/e510f71c72b5e2075dd39de0256febe765e5a3ab))
* use props safely ([b5f36e6](https://github.com/symphony09/running/commit/b5f36e68a7f44e4b7d7622d31d97db9498d78a74))



## [0.1.2](https://github.com/symphony09/running/compare/v0.1.1...v0.1.2) (2022-07-15)


### Bug Fixes

* fix engine bug ([2002dc8](https://github.com/symphony09/running/commit/2002dc8316caff9acf6122271b86e0fc36c8500a))
* fix missing node register ([fa8f0d8](https://github.com/symphony09/running/commit/fa8f0d8a0bc5be79980fa1b31cb19d7012d97fa4))


### Features

* support reuse nodes option ([b2d73fc](https://github.com/symphony09/running/commit/b2d73fccad4690988fc2741ca535eb5c23cae413))


### Performance Improvements

* add reuse node option on benchmark test ([2367530](https://github.com/symphony09/running/commit/23675302b703a36986a01eae30e9bb42e1e05b12))



## [0.1.1](https://github.com/symphony09/running/compare/4de82b0169977e9c8e930224b6e8af10202d968e...v0.1.1) (2022-07-04)


### Bug Fixes

* fix logic of new worker ([d91563a](https://github.com/symphony09/running/commit/d91563a320ceecef98bf665f430a17b0c17a714d))
* fix mod name ([5f559d9](https://github.com/symphony09/running/commit/5f559d98fe2c29ccc5eb11fd07b2c60a054e58a4))
* remove unnecessary reset  logic ([df4342d](https://github.com/symphony09/running/commit/df4342dca99ab2f12935a13e41052e957f5704df))


### Features

* add aspect cluster ([4e65fd5](https://github.com/symphony09/running/commit/4e65fd51bd75e39472d17628f5af8aa497258a09))
* add example ([77e374e](https://github.com/symphony09/running/commit/77e374e72b3bebcf85b4b5f66740d5558b0fd52b))
* add loop cluster ([64ef9e1](https://github.com/symphony09/running/commit/64ef9e18f7de3a7b4b1cf286b9495b4009fc7983))
* add merge cluster ([b4486d5](https://github.com/symphony09/running/commit/b4486d5ef7bc22d7f5dfa535c932d9db048144ec))
* add new plan option ([2847fdd](https://github.com/symphony09/running/commit/2847fdd4c73df721d52a76d93949b2ef281b56e7))
* add overlay state ([601f0db](https://github.com/symphony09/running/commit/601f0db6abafa61ef13d04c2d644c92e01fa38a4))
* add select cluster ([c9da9bc](https://github.com/symphony09/running/commit/c9da9bc695864549f9b5bf24b254f927b4038b13))
* add serial cluster ([5756239](https://github.com/symphony09/running/commit/57562392d8cb4e434daf933a4549c3b9584550ec))
* add simple node ([7c09db2](https://github.com/symphony09/running/commit/7c09db293b885de247358937abc8fca38421dafa))
* add some utils ([eb305da](https://github.com/symphony09/running/commit/eb305da975a4b05bea882a4f6f17f1dee26c39f0))
* add switch cluster ([3cb7c12](https://github.com/symphony09/running/commit/3cb7c12e4c9a17032a00cb939d37651a69441499))
* add worker pool ([56f7a27](https://github.com/symphony09/running/commit/56f7a27d35d30d9cd3cc90f5b45cdc47d7ce9f4b))
* support plan serialization ([2d4b1a1](https://github.com/symphony09/running/commit/2d4b1a1cf760204c03f02e2bbc077745f92660b2))
* support reverse link nodes ([f36ffb9](https://github.com/symphony09/running/commit/f36ffb9b0b807408ac596c3f3c8ee4f2462cee65))
* support set node name ([e36135b](https://github.com/symphony09/running/commit/e36135bbaed5ec181d41fb74fbe21248b1b43edd))
* support verify plan ([4de82b0](https://github.com/symphony09/running/commit/4de82b0169977e9c8e930224b6e8af10202d968e))
* support wrap node ([27c7949](https://github.com/symphony09/running/commit/27c7949ddc37247584fb978df8cc77b9c68d04bf))
* support wrap node ([9a67ed1](https://github.com/symphony09/running/commit/9a67ed13c0626ff1934b905db45d463cf05e2f72))


### Performance Improvements

* add context test ([55ba9b0](https://github.com/symphony09/running/commit/55ba9b029c405d36e3d0b130167152955799464b))
* add Global Func ([65c2431](https://github.com/symphony09/running/commit/65c24318600bc611f04b27f6ff8ca72fe80d212d))
* add new engine func ([bc2cde8](https://github.com/symphony09/running/commit/bc2cde8edd2c31f56ceb54d050d3416a3bd88b3f))
* add props test ([73fca8c](https://github.com/symphony09/running/commit/73fca8cebf12f4d1e9bceb2a93c2fc022a13505a))
* adjust cluster interface ([f64efad](https://github.com/symphony09/running/commit/f64efad9a3b9aa5c6d5a54f134d7103d0948ceb5))
* export cluster fields ([f667587](https://github.com/symphony09/running/commit/f6675876fb4fd9da4bd5a5e8be80006004948280))
* handle graph race ([d1ebae6](https://github.com/symphony09/running/commit/d1ebae6de3f43a713472fda567214a30b623ba82))
* handle node panic ([5e1af0b](https://github.com/symphony09/running/commit/5e1af0be209e01734e3bf3bbb672fa72f05002e3))
* handle node panic ([1d53aba](https://github.com/symphony09/running/commit/1d53aba778b8633219796462f628f6629567972d))
* handle node panic ([5e1f98e](https://github.com/symphony09/running/commit/5e1f98eb8b02a22a067f2607cb41abeb15cafb13))
* more safe to use plan ([321c990](https://github.com/symphony09/running/commit/321c990c09aeb5eeb7ef6bcc73c63b68f3d885c3))
* perf aspect cluster ([713f8f8](https://github.com/symphony09/running/commit/713f8f80ecfd4f387766e2d93b196134c51a1430))
* perf base node ([0594cda](https://github.com/symphony09/running/commit/0594cdab026c7c336f4b6489e07623469509164d))
* perf base reset ([0c66763](https://github.com/symphony09/running/commit/0c66763d698f62b819b940ba3c632aca02f85ab4))
* perf build node ([ff439b9](https://github.com/symphony09/running/commit/ff439b967df5e0811532356be07925b60ae0d2ab))
* perf build node ([ef4a939](https://github.com/symphony09/running/commit/ef4a9391e40b4db9dcfff5639985c42b2d640868))
* perf build node ([1197751](https://github.com/symphony09/running/commit/119775130f976627e7a6a52869b0a01368c7dd78))
* perf merge cluster ([87e4849](https://github.com/symphony09/running/commit/87e4849adeb6621664c9ddf8146c9f0db823ffb7))
* perf new plan function ([866d8bd](https://github.com/symphony09/running/commit/866d8bde44cf5633d90b1e92101006ed7fba7d49))
* perf node build function ([741f44f](https://github.com/symphony09/running/commit/741f44ff1465143dc286651516000edbd29822a7))
* perf props helper ([0880d7e](https://github.com/symphony09/running/commit/0880d7e0c900c1fa7ddb126046b1dc66e99dae5d))
* perf props interface ([19ae2d7](https://github.com/symphony09/running/commit/19ae2d7c300928ca1a4ebdd6219d87af6519e09c))
* perf run node ([63f0bee](https://github.com/symphony09/running/commit/63f0bee56c55562bf4e685207b13712741eb24e1))
* perf test ([a702b90](https://github.com/symphony09/running/commit/a702b90cf4e2922adfe4478e88feb5b151ac706a))
* perf test ([161b1ca](https://github.com/symphony09/running/commit/161b1cab2879f612c2170ab206277550bd75a952))
* support clear work pool ([d5ddaff](https://github.com/symphony09/running/commit/d5ddaff914325469b72bbb54f8726e9abf310628))
* support custom state builder ([43294fb](https://github.com/symphony09/running/commit/43294fbc3fe32c1d406f2cd5d8ebb0934589a132))
* support exec plan in strict mode ([bf7646b](https://github.com/symphony09/running/commit/bf7646b88772c061a9714c2d86361b5f28b6d798))
* support update plan ([3ed7b77](https://github.com/symphony09/running/commit/3ed7b77d2b71bc4deb415d5d239b8014f246ca56))



