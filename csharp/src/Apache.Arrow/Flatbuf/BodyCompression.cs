// <auto-generated>
//  automatically generated by the FlatBuffers compiler, do not modify
// </auto-generated>

namespace Apache.Arrow.Flatbuf
{

using global::System;
using global::System.Collections.Generic;
using global::Google.FlatBuffers;

/// Optional compression for the memory buffers constituting IPC message
/// bodies. Intended for use with RecordBatch but could be used for other
/// message types
internal struct BodyCompression : IFlatbufferObject
{
  private Table __p;
  public ByteBuffer ByteBuffer { get { return __p.bb; } }
  public static void ValidateVersion() { FlatBufferConstants.FLATBUFFERS_23_5_9(); }
  public static BodyCompression GetRootAsBodyCompression(ByteBuffer _bb) { return GetRootAsBodyCompression(_bb, new BodyCompression()); }
  public static BodyCompression GetRootAsBodyCompression(ByteBuffer _bb, BodyCompression obj) { return (obj.__assign(_bb.GetInt(_bb.Position) + _bb.Position, _bb)); }
  public void __init(int _i, ByteBuffer _bb) { __p = new Table(_i, _bb); }
  public BodyCompression __assign(int _i, ByteBuffer _bb) { __init(_i, _bb); return this; }

  /// Compressor library.
  /// For LZ4_FRAME, each compressed buffer must consist of a single frame.
  public CompressionType Codec { get { int o = __p.__offset(4); return o != 0 ? (CompressionType)__p.bb.GetSbyte(o + __p.bb_pos) : CompressionType.LZ4_FRAME; } }
  /// Indicates the way the record batch body was compressed
  public BodyCompressionMethod Method { get { int o = __p.__offset(6); return o != 0 ? (BodyCompressionMethod)__p.bb.GetSbyte(o + __p.bb_pos) : BodyCompressionMethod.BUFFER; } }

  public static Offset<BodyCompression> CreateBodyCompression(FlatBufferBuilder builder,
      CompressionType codec = CompressionType.LZ4_FRAME,
      BodyCompressionMethod method = BodyCompressionMethod.BUFFER) {
    builder.StartTable(2);
    BodyCompression.AddMethod(builder, method);
    BodyCompression.AddCodec(builder, codec);
    return BodyCompression.EndBodyCompression(builder);
  }

  public static void StartBodyCompression(FlatBufferBuilder builder) { builder.StartTable(2); }
  public static void AddCodec(FlatBufferBuilder builder, CompressionType codec) { builder.AddSbyte(0, (sbyte)codec, 0); }
  public static void AddMethod(FlatBufferBuilder builder, BodyCompressionMethod method) { builder.AddSbyte(1, (sbyte)method, 0); }
  public static Offset<BodyCompression> EndBodyCompression(FlatBufferBuilder builder) {
    int o = builder.EndTable();
    return new Offset<BodyCompression>(o);
  }
}


static internal class BodyCompressionVerify
{
  static public bool Verify(Google.FlatBuffers.Verifier verifier, uint tablePos)
  {
    return verifier.VerifyTableStart(tablePos)
      && verifier.VerifyField(tablePos, 4 /*Codec*/, 1 /*CompressionType*/, 1, false)
      && verifier.VerifyField(tablePos, 6 /*Method*/, 1 /*BodyCompressionMethod*/, 1, false)
      && verifier.VerifyTableEnd(tablePos);
  }
}

}
