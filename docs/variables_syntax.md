# Configuration variables syntax

The same variable can be retrieving using several providers (code, configuration file, command line, env variables ... see the [README.md](../README.md) file
to know the priority order)

For this reason it can be confusing on how to spell variables.

This documentation describes precisely each syntax for each type of provider.

Here is the struct used in all of this documentation :

```go
type NestedStruct struct {
	NestedVar  string
}
type ConfigStruct struct {
    StringVar  string
    IntVar     int
    NestedConf NestedStruct
}
```

## Examples
<table>
<tr>
<th>Provider</th>
<th>Description</th>
<th>Examples</th>
</tr>

<tr>
<td>Configuration file</td>
<td>
JSON, TOML, YAML files documentation does not force a developer to respect a given syntax, we do the same.

You can use in your configuration file the syntax of your choice between :
<ul>
<li>kebab-case</li>
<li>snake_case</li>
<li>camelCase</li>
</ul>
</td>
<td><pre><code class="language-json">{
    "string-var": "hello world",
    "int-var": 42,
    "nested-conf": {
        "nested-var": "Nested hello !"
    }
}</code>
                ----
<code class="language-json">{
    "string_var": "hello world",
    "int_var": 42,
    "nested_conf": {
        "nested_var": "Nested hello !"
    }
}</code>
                ----
<code class="language-json">{
    "stringVar": "hello world",
    "intVar": 42,
    "nestedConf": {
        "nestedVar": "Nested hello !"
    }
}</code></pre></td>
</tr>
<tr>
<td>Command line argument</td>
<td>
You can use only one syntax for command line parameters<br>
<ul><li>kebab-case</li></ul>
Notice the dot syntax to access nested variables.
</td>
<td><pre><code>
--string-var="hello world"
--int-var=42
--nested-conf.nested-var="Nested hello !"</code></pre></td>
</tr>
<tr>
<td>Environment variables</td>
<td>
You can use only one syntax for environment variable<br>
<ul><li>SCREAMING_SNAKE</li></ul>
Notice the underscore syntax to access nested variables.
</td>
<td><pre><code>
STRING_VAR="hello world"
INT_VAR=42
NESTED_CONF_NESTED_VAR="Nested hello !"
</tr>
</table>

