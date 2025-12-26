// Package tokenizer provides XML tokenization using Shape's tokenizer framework.
package tokenizer

// Token type constants for XML format.
// These correspond to the terminals in the XML grammar.
const (
	// Structural tokens
	TokenTagOpen       = "TagOpen"       // <
	TokenTagClose      = "TagClose"      // >
	TokenTagSelfClose  = "TagSelfClose"  // />
	TokenEndTagOpen    = "EndTagOpen"    // </
	TokenEquals        = "Equals"        // =

	// Content tokens
	TokenName          = "Name"          // Element/attribute names
	TokenString        = "String"        // "..." or '...' (attribute values)
	TokenText          = "Text"          // Text content between tags

	// Special sections
	TokenCDataStart    = "CDataStart"    // <![CDATA[
	TokenCDataEnd      = "CDataEnd"      // ]]>
	TokenCDataContent  = "CDataContent"  // Content inside CDATA

	// Declaration/Processing Instructions
	TokenXMLDeclStart  = "XMLDeclStart"  // <?xml
	TokenPIStart       = "PIStart"       // <?
	TokenPIEnd         = "PIEnd"         // ?>

	// Comments
	TokenCommentStart  = "CommentStart"  // <!--
	TokenCommentEnd    = "CommentEnd"    // -->
	TokenCommentContent = "CommentContent" // Comment text

	// Special token
	TokenEOF           = "EOF"           // End of file
)
