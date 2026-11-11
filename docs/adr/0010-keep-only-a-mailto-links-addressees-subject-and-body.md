# 0010 — Keep only a mailto link's addressees, subject and body

- **Status:** Accepted
- **Date:** 2026-11-11

## Context

[ADR-0008](0008-open-behind-allow-lists-and-let-editors-keep-their-keys.md)
lets `Opener` open mailto links by default, and it passed them to the
system's mail client whole. A mailto link carries header fields after its
`?`, and some mail clients act on fields other than the message's text:
`attach=` or `attachment=` attaches a local file, as `/etc/passwd` or a
private key, and other headers set where a reply goes or what the message
answers. A link in a message or a note could then make a message that sends
a file the user never chose
([frostyard/frostui#12](https://github.com/frostyard/frostui/issues/12),
from the review of frostyard/frostmail#223). An app that is itself the
system's mailto handler receives such links from other programs and needs
the same rule.

RFC 6068 writes a mailto link as `mailto:` [addresses] [`?` field `=` value
*(`&` field `=` value)]. It has no fragment. Mail clients differ in how they
read the rest: a client may decode a field's name before it compares it, or
read a `#` and what follows as more of the link's fields.

## Decision

**What survives.** A mailto link keeps its address and only the fields named
`to`, `cc`, `bcc`, `subject` and `body`, in any ASCII case, as written. Every
other field is removed, and the link opens without it: a field is dropped,
never a reason to refuse the link. `SanitizeMailto` exports the rule, and
`Opener.CheckURL` and `OpenURL` apply it to every mailto link, whatever
`Schemes` lists. An opener still refuses a link that names nothing (ADR-0008),
so a link left with no address and no field is refused as `mailto:` is;
`SanitizeMailto` returns it as `mailto:`.

**How.** The rule decodes nothing, and so cannot read a name differently
from a mail client:

- Fields are split on `&`, and a field's name ends at its first `=`. The
  name must be one of the five, letter for letter: an encoded name, as
  `%61ttach` or even `%74o`, is dropped.
- A field with no `=` is not a field (RFC 6068) and is dropped, kept name or
  not. An empty field, as in `&&`, goes too. `subject=` is kept.
- A kept field passes byte for byte: `+` stays `+` and `%20` stays `%20`, and
  a malformed escape, as `100%`, stays as written. The mail client reads
  them as it would without the rule.
- A repeated field is kept every time, in order: `to=a&to=b` names two
  people, and for a repeated subject or body the client chooses as it would
  without the rule.
- An address before the `?` and a `to=` field both stay: RFC 6068 makes both
  recipients.
- The fragment is removed: mailto has none, and a client that does not
  split one off would read `#&attach=…` as a field.
- A link with nothing left after its address loses its `?`:
  `mailto:a@example.org?attach=x` becomes `mailto:a@example.org`.

`SanitizeMailto` refuses, with an error wrapping `ErrNotAllowed`, what is not
a mailto link, or holds spaces or control characters, or does not parse, as
`CheckURL` does.

## Consequences

- A mailto link from outside an app can name recipients, a subject and a
  body, and nothing else: no attachment, no reply-to, no other header.
- An app that handles mailto links calls `SanitizeMailto` before it fills a
  message, and keeps no copy of the rule.
- A link whose kept field is encoded, as `%74o=`, loses that field. Such
  links are rare, and the rule stays a comparison of names as written.
- The rule does not look inside a kept value. A value holding an encoded
  line break or `&` reaches the mail client as text, which RFC 6068's
  security considerations leave to the client.
- A new kept field is a new ADR, as it widens what a link may ask of a mail
  client.

## Alternatives considered

- **Refusing a link with another field:** the link's recipients and text are
  still useful, and a refusal tells the user nothing they can act on.
- **Decoding names before comparing them:** a decoder that differs from the
  mail client's, or double encoding (`%2561ttach`), could let a field
  through. Comparing names as written keeps only what any client reads as
  one of the five.
- **Keeping the fragment, as the link's text:** a client that reads `#` as
  data would take the fields after it.
- **Splitting on `;` as well as `&`:** RFC 6068 separates fields with `&`
  alone, and `;` in a value is text.
- **Keeping only the first of a repeated field:** it would lose recipients,
  and protects nothing the five names do not already allow.

## References

- Shapes: [design: frostui](../design/frostui.md#opener)
- Builds on: [ADR-0008](0008-open-behind-allow-lists-and-let-editors-keep-their-keys.md)
