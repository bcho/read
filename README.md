# Read

[![Build Status](https://travis-ci.org/bcho/read.svg)](https://travis-ci.org/bcho/read)

## Usage

### Record read article

```
ME: /read http://example.com/an-article
          This article describe a cool technology... / nope. (just leave comment blank)
read: Copy that! New link http://example.com/an-article added.
```

### Get read articles in a period

```
ME: /stats last week
read: You read 3 articles during 2016-01-01 ~ 2016-01-07

      http://example.com/an-article

      This article describe a cool technology...

      ...
```

### Add a bookmark

```
ME: /bookmark http://example.com/another-article
read: Roger that! New link http://example.com/another-article added.
```

### Get a random bookmark

```
ME: /random
read: http://example.com/another-article / No more bookmarks, nice!
```
