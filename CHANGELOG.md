# Changelog

## [2.0.0](https://github.com/andreoliwa/logseq-go/compare/v1.4.0...v2.0.0) (2026-10-18)


### ⚠ BREAKING CHANGES

* **content:** use regular properties with editable source styles
* **markdown:** default to typed logbooks with preserved source
* **pages:** expose the upstream page contract with pre-blocks
* **markdown:** make source preservation opt in
* **content:** separate status setters from task transitions
* **content:** align priority APIs and positional parsing
* **module:** adopt the v2 import path

### Features

* **content:** align priority APIs and positional parsing ([0aaa5e3](https://github.com/andreoliwa/logseq-go/commit/0aaa5e37a10c8a9562656184788cb119590eb90d))
* **content:** separate status setters from task transitions ([f58d3cc](https://github.com/andreoliwa/logseq-go/commit/f58d3cc0426d0a6cc15a7c2cd755fd352436893c))
* **content:** use regular properties with editable source styles ([0b659cd](https://github.com/andreoliwa/logseq-go/commit/0b659cdfe939c5461b7e4fbdeff50f727df23f88))
* **markdown:** default to typed logbooks with preserved source ([17a19e5](https://github.com/andreoliwa/logseq-go/commit/17a19e5b1ed5c37c48072c7effcdfcc69304942d))
* **markdown:** make source preservation opt in ([c3e6e6a](https://github.com/andreoliwa/logseq-go/commit/c3e6e6a439afee3ba5822596d98eba673b7d506a))
* **module:** adopt the v2 import path ([0f3a340](https://github.com/andreoliwa/logseq-go/commit/0f3a340e826e7dc9b90fafd93804bd55c25ca867))
* **pages:** expose the upstream page contract with pre-blocks ([67513ee](https://github.com/andreoliwa/logseq-go/commit/67513ee16afa807cbbc9976114f2cab3d4374633))
* **upstream:** report cached upstream changes ([1b03002](https://github.com/andreoliwa/logseq-go/commit/1b03002cd91c811ad31de83878e98d6fadb6a0ee))


### Bug Fixes

* **markdown:** align canonical block and header spacing ([8e3a870](https://github.com/andreoliwa/logseq-go/commit/8e3a8702ff38a021fe55d9dbf346ba518dd8524f))
* **markdown:** align canonical escaping with upstream ([2fdfd11](https://github.com/andreoliwa/logseq-go/commit/2fdfd11b16b40665a6840a5d6795af741dda6192))
* **markdown:** render canonical formatting without losing source style ([f651ea8](https://github.com/andreoliwa/logseq-go/commit/f651ea8b12b2971624af7d2912b80af9b9989293))
* **transaction:** preserve page endings and skip identical writes ([8b01523](https://github.com/andreoliwa/logseq-go/commit/8b01523052f2519a7f056f0fb0593433843bd7c8))
* **upstream:** support Windows checks and test fixtures ([159d96e](https://github.com/andreoliwa/logseq-go/commit/159d96ea17c17401b468e3baa7f1e9c5d0d34fc0))

## [1.4.0](https://github.com/andreoliwa/logseq-go/compare/v1.3.0...v1.4.0) (2026-10-17)

### Bug Fixes

- Property blocks at start of page no longer break indexing ([9feed0b](https://github.com/andreoliwa/logseq-go/commit/9feed0ba5dad799ed8130908814aeba43c36462f))
- Saving a page keeps the file's trailing newline ([d27ca1a](https://github.com/andreoliwa/logseq-go/commit/d27ca1a52856a1444fdddb9e4109ec4fe49cda47))
- Pages that mix properties and prose without bullets keep their bare form ([900423e](https://github.com/andreoliwa/logseq-go/commit/900423e1d22f9dfebee4f9f504a302cfc2e04625))
- Closing a graph while changes are being indexed no longer crashes ([c04ef84](https://github.com/andreoliwa/logseq-go/commit/c04ef847f1fe4760991e04d95d3dd8fa0248b2e9))
- Empty query macro no longer panics when parsing ([cf40a4c](https://github.com/andreoliwa/logseq-go/commit/cf40a4c6624c83a7419116f88fc4277cd8ed49ab))
- Macro with three curly braces and a missing brace no longer panics ([bf586bd](https://github.com/andreoliwa/logseq-go/commit/bf586bd6600bd8c55a245578f5b66c3aaa32f440))
- Journals resolve to the right day in zones behind UTC ([d57a7c8](https://github.com/andreoliwa/logseq-go/commit/d57a7c85009c0289ec16f4f37eab30da8e7eb8c8))
- Journals opened in a transaction are saved to the file they came from ([6e9b67e](https://github.com/andreoliwa/logseq-go/commit/6e9b67e7fa05a42e35e363a1472f9e2c55010761))
- Hash followed by a space is no longer parsed as a tag ([b05dbd0](https://github.com/andreoliwa/logseq-go/commit/b05dbd03bc48e4f774b12cc06639896ab6f02a18))
- Task marker on its own no longer written twice ([20476f4](https://github.com/andreoliwa/logseq-go/commit/20476f4939184558c2396c090658fb4524a2f9da))
- Code spans containing backticks are written so they can be read back ([0a48197](https://github.com/andreoliwa/logseq-go/commit/0a4819775fc92562edcc9ac1c2a2b0661753d012))
- Full weekday and month names in date formats convert correctly ([e525b89](https://github.com/andreoliwa/logseq-go/commit/e525b89b1bbe9f967ccf15d0c1b533e4c6db0877))
- Pages containing raw text can be saved and searched ([7f0178a](https://github.com/andreoliwa/logseq-go/commit/7f0178a360b5991dc5250461f79af74e4c85a1b1))
- Setting children detaches all of the previous ones ([f58863f](https://github.com/andreoliwa/logseq-go/commit/f58863fe72f754f3b0f9294a118357ca72949d52))
- Closing a watcher no longer stops the index being updated ([1a1d356](https://github.com/andreoliwa/logseq-go/commit/1a1d356a30640b5fa262dee9236a44fc8ac6027a))
- Nodes that are not allowed as children are rejected everywhere ([e8ed320](https://github.com/andreoliwa/logseq-go/commit/e8ed320d391769408d5fa6c8c74e252eb95fced9))
- Unknown list marker no longer kept as the marker of the list ([e30e45b](https://github.com/andreoliwa/logseq-go/commit/e30e45b108bc2a67c1e76cc4c15c5e33ac093c9c))
- Closing a watcher twice no longer panics ([92d5594](https://github.com/andreoliwa/logseq-go/commit/92d55941a04351f00533f0ffa8efccca3f4d393b))
- Model task priorities and stop escaping their brackets ([f075595](https://github.com/andreoliwa/logseq-go/commit/f0755952e0b5d2d8c03431ddf9bf8842b544e5de))
- Keep the whole value when parsing properties ([b8c24e4](https://github.com/andreoliwa/logseq-go/commit/b8c24e43f3f1bfad52069f50c651aaca0493fd03))
- Write the ordinal day in journal titles so they match Logseq ([b1279c3](https://github.com/andreoliwa/logseq-go/commit/b1279c3e16142310d17e33f01158a030d2139944))
- Skip Markdown files that are not pages or journals when syncing ([90cfe6d](https://github.com/andreoliwa/logseq-go/commit/90cfe6d31bde2a855a9ac2cfb78287c325e16fb8))
- **release:** Include upstream commits in proposed changelogs ([5c2df17](https://github.com/andreoliwa/logseq-go/commit/5c2df173e2c0aaa7f6fc08731d45c75533e71c40))
- **release:** Format compact entries with GitHub usernames ([97354e0](https://github.com/andreoliwa/logseq-go/commit/97354e0cbc09a9472c2d8fce7b148814b846d20e))
- **release:** Omit author labels from changelog entries ([0e5c63c](https://github.com/andreoliwa/logseq-go/commit/0e5c63c5c89ca0d004c655fa04cc771f6d6ed080))

### Documentation

- **release:** Document direct master development ([08d8888](https://github.com/andreoliwa/logseq-go/commit/08d88883def055b374d39fb5e53d3ca0bebe0fc5))

### Features

- **release:** Add event-scoped release notes ([1653c3a](https://github.com/andreoliwa/logseq-go/commit/1653c3a29898a12faca592abb18cf2d874bed9bf))
- **release:** Adopt Release Please for reviewed releases ([29a0850](https://github.com/andreoliwa/logseq-go/commit/29a0850ca7e83c9ac48054f788bf24fe9c151310))
- Model SCHEDULED and DEADLINE dates on tasks ([1e6f25d](https://github.com/andreoliwa/logseq-go/commit/1e6f25d4f5809ecb96a1c994bf398ce0c9906909))
- Model clock and state change entries in logbooks ([a46b150](https://github.com/andreoliwa/logseq-go/commit/a46b15003d436d89f318a86b3b2980e898b45652))
- Rename and delete pages, with reference rewriting ([0235017](https://github.com/andreoliwa/logseq-go/commit/02350177230d2416b7b55e6723d24c81f7686ff1))
- Resolve blocks by their id ([fff8fd4](https://github.com/andreoliwa/logseq-go/commit/fff8fd4e44ae3e768f4ff242c7dbb87192006a4d))
- Find the linked references of a page ([a969334](https://github.com/andreoliwa/logseq-go/commit/a969334c95112cc4907f78438a11f46136ffc1de))
- Resolve pages by their aliases ([18ed5a4](https://github.com/andreoliwa/logseq-go/commit/18ed5a4d20a0bf53fbd98c04b2633514e09e7c86))
- Model the namespaces that pages are in ([19946aa](https://github.com/andreoliwa/logseq-go/commit/19946aa98efcb86418004f9134097d8156678fd8))
- Rename the pages in a namespace with the page ([673c4d2](https://github.com/andreoliwa/logseq-go/commit/673c4d2e26bfd22c5cc9a92a8126c95c346fa9bc))
- Read and use more of the Logseq config ([af06822](https://github.com/andreoliwa/logseq-go/commit/af0682241e73375ab3d4a30ec02cd5899b103f56))
- Honor :with-second-support? when writing and parsing logbooks ([a6c1e92](https://github.com/andreoliwa/logseq-go/commit/a6c1e92d9a75cc44a0806c91a625cedbfd8cf506))
- Apply :property/separated-by-commas and :ignored-page-references-keywords when parsing ([241c3ed](https://github.com/andreoliwa/logseq-go/commit/241c3edabe4418f5aeae7061ab1026f52f749b41))
- Read and write ^^highlights^^ ([a039d26](https://github.com/andreoliwa/logseq-go/commit/a039d26909f754a8026f60e1cbef8363c52e15b9))
- Read and write GFM tables ([7ab5f93](https://github.com/andreoliwa/logseq-go/commit/7ab5f93f1c50767ffdb5a69171d5aead0f3d63a9))
- Read and write LaTeX formulas ([cd5a77c](https://github.com/andreoliwa/logseq-go/commit/cd5a77c65f8671ae2281c67a61af86855e85b262))
- Read and write footnotes ([1cc097e](https://github.com/andreoliwa/logseq-go/commit/1cc097e4d1e36d5f28a02d24d4c834c005e3d194))
- **logseq:** Integrate upstream capabilities compatibly ([8b73b27](https://github.com/andreoliwa/logseq-go/commit/8b73b27b344b416a305a90ca1a53d53087bba49c))

## [1.3.0](https://github.com/andreoliwa/logseq-go/compare/v1.2.0...v1.3.0) (2026-10-16)

### Features

- **content:** add leading time recognition and normalization ([6f026fe](https://github.com/andreoliwa/logseq-go/commit/6f026fee3095c1adc419441a759eb5e6f374bf1b))
- **markdown:** expose raw outline source ranges ([9b8dc9a](https://github.com/andreoliwa/logseq-go/commit/9b8dc9a7f479d80a623d9511fa08857540fc38db))

## [1.2.0](https://github.com/andreoliwa/logseq-go/compare/v1.1.0...v1.2.0) (2026-03-29)

### Features

- String constants for all Task statuses ([3ca90e8](https://github.com/andreoliwa/logseq-go/commit/3ca90e87801f765e5643541d37c40902db1ecae1))

## [1.1.0](https://github.com/andreoliwa/logseq-go/compare/v1.0.1...v1.1.0) (2026-03-29)

### Bug Fixes

- DeletePage now correctly handles relative paths ([494dafc](https://github.com/andreoliwa/logseq-go/commit/494dafcce4f5d6f25960534facfb98c0fedd9f50))
- Ensure journal date is truncated to midnight to avoid timezone issues ([69cfa37](https://github.com/andreoliwa/logseq-go/commit/69cfa3737dd8a535fe44d6f6298b5abf55b75a98))
- Graph watcher reported delete page when page was just updated ([89bc477](https://github.com/andreoliwa/logseq-go/commit/89bc4771c9f4cbd652a36b9ff1cde9b5d60e9ad8))
- update module paths and resolve races after upstream sync ([15b0396](https://github.com/andreoliwa/logseq-go/commit/15b039671a32e9264a03a7c141a020bca3d66f3a))
- Various small search fixes and tests ([8e42a0b](https://github.com/andreoliwa/logseq-go/commit/8e42a0b6dbedcab965148e926032beb56b702939))

### Features

- Add TitlePartiallyMatches for index ([e18debd](https://github.com/andreoliwa/logseq-go/commit/e18debd9a93eb3edcdf14ad27a196dcc51ee09ae))
- AsString helper added for nodes ([7379045](https://github.com/andreoliwa/logseq-go/commit/7379045e065c688042428ab30fb1a1371393ee35))
- serialize content to Markdown ([7bfceab](https://github.com/andreoliwa/logseq-go/commit/7bfceabfdd4436dac3f5c63f128ef6fc63142caf))

## [0.4.1](https://github.com/andreoliwa/logseq-go/compare/v0.4.0...v0.4.1) (2025-11-01)

### Bug Fixes

- parse CLOCK in local timezone ([7865e02](https://github.com/andreoliwa/logseq-go/commit/7865e0241db03f3ab6704e237f3ab342a73a1256))

## [0.4.0](https://github.com/andreoliwa/logseq-go/compare/v0.3.1...v0.4.0) (2025-11-01)

### Features

- change any task status to TODO, adjusting the logbook ([50cda14](https://github.com/andreoliwa/logseq-go/commit/50cda14bc5b1395cf1bc0373b62383c7976e3d43))

## [0.3.1](https://github.com/andreoliwa/logseq-go/compare/v0.3.0...v0.3.1) (2025-11-01)

### Bug Fixes

- preserve previous line when replacing nodes ([11425b6](https://github.com/andreoliwa/logseq-go/commit/11425b64a6b8968245ef7aa68324f814b5abcb74))

## [0.3.0](https://github.com/andreoliwa/logseq-go/compare/v0.2.2...v0.3.0) (2025-11-01)

### Features

- parse and output the alias property ([c9ef91e](https://github.com/andreoliwa/logseq-go/commit/c9ef91e7d0b0d35b67989fd58afb077faebf0981))

## [0.2.2](https://github.com/andreoliwa/logseq-go/compare/v0.2.1...v0.2.2) (2025-08-23)

### Bug Fixes

- don't escape underscores in Markdown links ([80e2855](https://github.com/andreoliwa/logseq-go/commit/80e28557f2a629e2c5781345da31cac0905bafdd))

## [0.2.1](https://github.com/andreoliwa/logseq-go/compare/v0.2.0...v0.2.1) (2025-07-30)

### Bug Fixes

- don't escape parenthesis in valid URLs ([5214141](https://github.com/andreoliwa/logseq-go/commit/521414142d91bd2b6aec22f3ada85db19b0dd660))
- keep original line breaks in block quotes ([33ad53b](https://github.com/andreoliwa/logseq-go/commit/33ad53baa748f84d0f3b8fb1893ae57fd61b8857))
- preserve tabs and spaces in code blocks ([53dbc29](https://github.com/andreoliwa/logseq-go/commit/53dbc290c1b8c2d88352d275b4426137e14195a1))

## [0.2.0](https://github.com/andreoliwa/logseq-go/compare/v0.1.5...v0.2.0) (2025-07-28)

### Features

- new content type to parse priorities ([2564c23](https://github.com/andreoliwa/logseq-go/commit/2564c23ef23a697872d5c267d30e1566b693894e))

## [0.1.5](https://github.com/andreoliwa/logseq-go/compare/v0.1.4...v0.1.5) (2025-07-26)

### Bug Fixes

- don't escape brackets, underscores and stars ([81ca657](https://github.com/andreoliwa/logseq-go/commit/81ca657d01f4a64233934b64d3a29eb01a766b82))
- preserve tabs and spaces in the beginning of empty lines ([aa0bb67](https://github.com/andreoliwa/logseq-go/commit/aa0bb67c2846a4bcd4f801a0199f9fe29ed48e10))

## [0.1.4](https://github.com/andreoliwa/logseq-go/compare/v0.1.3...v0.1.4) (2025-07-20)

### Bug Fixes

- handle multiple cases of links with hashtags ([ea1455b](https://github.com/andreoliwa/logseq-go/commit/ea1455b6fdb19a7e1201f6bea548e60f7711bd21))
- preserve list item numbers inside block quotes ([a1362e1](https://github.com/andreoliwa/logseq-go/commit/a1362e19b913db8abcab7872e86483466c6165d2))

## [0.1.3](https://github.com/andreoliwa/logseq-go/compare/v0.1.2...v0.1.3) (2025-07-19)

### Bug Fixes

- handle properties with spaces in the value ([63bf0ed](https://github.com/andreoliwa/logseq-go/commit/63bf0ed20fa38afb2e278bf8c2c106de38ea533f))
- header output ([95723ba](https://github.com/andreoliwa/logseq-go/commit/95723ba1081e8c6b2a4e3b9a576a72465d9756a2))
- keep original spacing of block quotes ([5af6e23](https://github.com/andreoliwa/logseq-go/commit/5af6e23c7766452b254d823fdf331044676d463e))
- parsing of escaped strikethrough ([796f351](https://github.com/andreoliwa/logseq-go/commit/796f3514768e31507db585b66f0b3a18b41b6489))
- preserve block quote signs and spacing ([0907be2](https://github.com/andreoliwa/logseq-go/commit/0907be261c0d6c9c2ac50d9a775c0384fcc0aca3))
- preserve original italics char (underscore or star) ([ca79289](https://github.com/andreoliwa/logseq-go/commit/ca792891fb9f691c6ef206c52fa697cfcce42b5c))
- task status must be uppercase only ([16ff80c](https://github.com/andreoliwa/logseq-go/commit/16ff80c64fd52dcfb63a17fec1d2d0399fbd6403))
- thematic break as a list item ([63ba750](https://github.com/andreoliwa/logseq-go/commit/63ba750c04d62fcf17edb168bbdd54be13bc6d04))
- treat extra dashes as literals, not bullets ([739e769](https://github.com/andreoliwa/logseq-go/commit/739e769dda7008975f7f10105024b35eaf3ff1e0))

## [0.1.2](https://github.com/andreoliwa/logseq-go/compare/v0.1.1...v0.1.2) (2025-02-24)

### Bug Fixes

- bump all package versions ([48dc192](https://github.com/andreoliwa/logseq-go/commit/48dc19275416ef0c35ba0a8304b30a3f9192629e))

## [0.1.1](https://github.com/andreoliwa/logseq-go/compare/v0.1.0...v0.1.1) (2025-02-24)

### Bug Fixes

- retract version v1.0.0 published by mistake ([3e06bdb](https://github.com/andreoliwa/logseq-go/commit/3e06bdb39d708612009cf913a2fbcdd691ce99ca))

## [0.1.0](https://github.com/andreoliwa/logseq-go/compare/v0.0.0...v0.1.0) (2025-02-23)

### Bug Fixes

- Adding journal blocks should skip properties when looking for times ([9dfe949](https://github.com/andreoliwa/logseq-go/commit/9dfe9493dd5d1e5d22e0e89ee32328bbcbb47615))
- Adjust some property parsing cases ([3816ce8](https://github.com/andreoliwa/logseq-go/commit/3816ce8090b77bb37b79fe1a68b9cf7ff5d994f9))
- Block properties can now be fetched correctly ([2308dcd](https://github.com/andreoliwa/logseq-go/commit/2308dcdcc2ba86a06aa070e91cd98734b22f2a78))
- Code blocks no longer grow an extra newline after parse/write ([aefd884](https://github.com/andreoliwa/logseq-go/commit/aefd88469de9caca9f0ab9a7a9e5bfab9c68c12f))
- Don't add extra newlines when writing properties with paragraph siblings ([2e1b2aa](https://github.com/andreoliwa/logseq-go/commit/2e1b2aac35833a2193b726b3d0b8899099d2ccba))
- don't escape potential Markdown in text nodes ([3b9f58b](https://github.com/andreoliwa/logseq-go/commit/3b9f58bed6b72efb87f8488c42d8bf7629e04b8d))
- Fix decoding of URL encoded characters ([3922bc2](https://github.com/andreoliwa/logseq-go/commit/3922bc23b3c17410df13db57fa3651b83ce44934))
- Fix handling of non-existent pages and journals ([7d743af](https://github.com/andreoliwa/logseq-go/commit/7d743af67b0f7d3573a33cfa5e286c07a3d46870))
- Fix issue with parsing advanced commands ([7de762e](https://github.com/andreoliwa/logseq-go/commit/7de762e5d0c452596d215a992c0fdef2fd0a2a88))
- Fix panic when writing top-level blocks ([504b4e1](https://github.com/andreoliwa/logseq-go/commit/504b4e13ddcd5b0e2871deee91edb4db908731ef))
- Index page refs and hashtags as keyword fields ([60b2767](https://github.com/andreoliwa/logseq-go/commit/60b2767eb98a7d0e7e85de4e329f0b87114b1be1))
- **markdown:** Fix naming of AdvancedCommand ([22bb448](https://github.com/andreoliwa/logseq-go/commit/22bb448fee2f8e3139879629d6003d2925dfdafa))
- **markdown:** Fix output of non query advanced commands ([e7c2261](https://github.com/andreoliwa/logseq-go/commit/e7c2261031ad6fd15c855cf3dd24cec35166d51e))
- Parse dash without trailing space as a block ([3ad008b](https://github.com/andreoliwa/logseq-go/commit/3ad008ba88ed765ed1b9ea9c76c30a836307f33f))
- Parse properties in blocks ([32ce7a7](https://github.com/andreoliwa/logseq-go/commit/32ce7a74d70e489fe157c3862b2de3f48afc2c07))
- Properties now support non-text values and handle line breaks better ([075c86c](https://github.com/andreoliwa/logseq-go/commit/075c86c81f3b879a2faa00b63adf9b672838ae50))
- Remove debug log ([c41a73c](https://github.com/andreoliwa/logseq-go/commit/c41a73c6e8d8aba3e8c97271ceceee4f120ce153))
- Separate emphasis siblings with a space ([aa0045e](https://github.com/andreoliwa/logseq-go/commit/aa0045e17d7054665ab593ed71bb55c740fd3cf3))
- Separate strong siblings with a space ([ff4f6ab](https://github.com/andreoliwa/logseq-go/commit/ff4f6ab3edb34b137021e6187fbcc9fa86200fbb))
- Skip pages that do not exist, and include title ([2583598](https://github.com/andreoliwa/logseq-go/commit/2583598d075839bc31dcfca72a7135ae32488279))
- Sync index after indexing changes ([d9fc066](https://github.com/andreoliwa/logseq-go/commit/d9fc066976314275985557d653df981af4386936))
- truncate journal date ([db509e3](https://github.com/andreoliwa/logseq-go/commit/db509e37476f8d7ed05ddbad187b4c0e32b3f944))
- Using strikethrough text could cause a panic ([65437fc](https://github.com/andreoliwa/logseq-go/commit/65437fc1205877f335e2410c396b13041432559b))

### Features

- Add basic support for logbooks ([13f0a7d](https://github.com/andreoliwa/logseq-go/commit/13f0a7d595a11222e0dc83a996eb928430771c9c))
- Add full strikethrough support ([6156e09](https://github.com/andreoliwa/logseq-go/commit/6156e09e809aedc3158915afddd6c7b9946ea6c2))
- Add method to get block ID ([e6602a6](https://github.com/andreoliwa/logseq-go/commit/e6602a6773aae95336370766e4a6686538d756c6))
- Add node type for Cloze macro ([267cc2e](https://github.com/andreoliwa/logseq-go/commit/267cc2e2cd10e21ed6b47118c986d526c7481f71))
- Add ParseBlock and ParseNodes functions ([0344de6](https://github.com/andreoliwa/logseq-go/commit/0344de63f6ad65ec3f62f9494681e52cd9200ca3))
- Add parsing and output of task markers ([35b86cf](https://github.com/andreoliwa/logseq-go/commit/35b86cff32ea05470c24742f6ef8bc9f11914575))
- Add support for searching blocks ([4460ee3](https://github.com/andreoliwa/logseq-go/commit/4460ee3ee06fd0adde9a7d8c3c2e485d36761d1f))
- Align block indentation with Logseq ([e817fdf](https://github.com/andreoliwa/logseq-go/commit/e817fdf7a080d32324d6e447e871d9d2e3f2eb91))
- Automatically add paragraphs to blockquote children ([5aa3bb0](https://github.com/andreoliwa/logseq-go/commit/5aa3bb00445f973ece6dcfac80183fd5ea8fc6e8))
- Basic support for indexing and searching journals and pages ([bd9e1be](https://github.com/andreoliwa/logseq-go/commit/bd9e1be1dc043a8cfcdb1ee06d4cc571e1a5af30))
- Block WithID now only sets the ID and returns the block ([30b1ac3](https://github.com/andreoliwa/logseq-go/commit/30b1ac3c4facace8d73aab9dc55ebcaa40a66465))
- export journal path function ([154c5f3](https://github.com/andreoliwa/logseq-go/commit/154c5f32541cfd1d6396866d2acc7de45b1a08a8))
- Improve handling of properties ([4980504](https://github.com/andreoliwa/logseq-go/commit/49805045e849fa53d91fb5d5c07a3e2feb826aea))
- Improve macro parsing and output ([e2bbe8d](https://github.com/andreoliwa/logseq-go/commit/e2bbe8d2b2807781a0d779f3087020156c000b3f))
- Improve parsing of macros ([79075b1](https://github.com/andreoliwa/logseq-go/commit/79075b1cbe61ae778ec4ea6d41e2276320cd19a2))
- Improve search API ([6794a03](https://github.com/andreoliwa/logseq-go/commit/6794a03e9153dd2bba2c07586fbe201437edf8eb))
- Index content of notes for full-text search ([8e1c843](https://github.com/andreoliwa/logseq-go/commit/8e1c8439bf8a7d4201aaab77f7d39aac8caa974c))
- Index links in pages and blocks ([f92a2d4](https://github.com/andreoliwa/logseq-go/commit/f92a2d4ad452b86d112e218301a38b9ec8727a4c))
- Index non-toplevel blocks ([89c1ccb](https://github.com/andreoliwa/logseq-go/commit/89c1ccb3e654d6d569dfac00f2bca247ac8900f8))
- Introduce Query, PageEmbed and BlockEmbed nodes ([eb00409](https://github.com/andreoliwa/logseq-go/commit/eb004099bab47048e26bf6413aa0e1172411dc0c))
- Introduce transactions for editing content ([42aca55](https://github.com/andreoliwa/logseq-go/commit/42aca5584016b179c592d3d5e2f06c62daea32ad))
- **markdown:** Add block reference parsing ([7b3af8e](https://github.com/andreoliwa/logseq-go/commit/7b3af8ecbff2d585236036bc018eae1e5bb41610))
- **markdown:** Add initial parsing of macros ([92dcd28](https://github.com/andreoliwa/logseq-go/commit/92dcd282967eae5ea312dfb406c4441290993ce9))
- **markdown:** Cleaner auto link output and parsing ([a800020](https://github.com/andreoliwa/logseq-go/commit/a800020cdba43f6c67f5412b10108da0c03f5736))
- **markdown:** Parse advanced commands ([f6f0c31](https://github.com/andreoliwa/logseq-go/commit/f6f0c31b5b33ec39eadff2bff9d1dfbcc59f8c5f))
- **markdown:** Support specific QueryCommand ([fc6a4c8](https://github.com/andreoliwa/logseq-go/commit/fc6a4c82211c1d63d4ff5c139237faf52ba715b3))
- Match using AND when searching for text ([10bb5cc](https://github.com/andreoliwa/logseq-go/commit/10bb5ccb5a0c274238ce160af176e3a5edaf1154))
- More helpers for content nodes ([31e3992](https://github.com/andreoliwa/logseq-go/commit/31e39926bbe62640363e383b3fd500c31f293470))
- Move journal block adding to Transaction ([092eeeb](https://github.com/andreoliwa/logseq-go/commit/092eeeb2962ab693875453ddf1c7bce3d23c0292))
- Open a transaction file via path ([32088a5](https://github.com/andreoliwa/logseq-go/commit/32088a5a9491c1b4926b1411d982f6130d5e0315))
- Open files directly via path ([1fa8678](https://github.com/andreoliwa/logseq-go/commit/1fa8678c1940c920982e5bad98bec17fa72cd3b9))
- Parse a task status from a string ([481dcf3](https://github.com/andreoliwa/logseq-go/commit/481dcf381a82177554d048137db74f088ff7c927))
- Parse and output markers of lists ([961f240](https://github.com/andreoliwa/logseq-go/commit/961f2409c2f35690b48e8b45a92945a063103001))
- Rename InsertBefore/After to InsertChildBefore/After ([3ad5d6e](https://github.com/andreoliwa/logseq-go/commit/3ad5d6e37801a6712b06604878966365f8efe0d1))
- Support for watching for page changes ([4727f2b](https://github.com/andreoliwa/logseq-go/commit/4727f2b2f0573bee000b2b8e257e52af8794bafa))
- Track if previous line is blank to keep Markdown output close to input ([f303198](https://github.com/andreoliwa/logseq-go/commit/f3031985c0149de57b6174941aec18b56d5fcf7d))
- Transactions now support searching ([d7baddf](https://github.com/andreoliwa/logseq-go/commit/d7baddf0f2290bdf998e7419f1d17a5e152371a9))
- Watch for changes to graph when indexing is active ([db35275](https://github.com/andreoliwa/logseq-go/commit/db352750709801d4cda19a5e4ec6b68fc6c99e38))
