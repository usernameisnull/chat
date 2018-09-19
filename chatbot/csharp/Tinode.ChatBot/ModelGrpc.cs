// <auto-generated>
//     Generated by the protocol buffer compiler.  DO NOT EDIT!
//     source: model.proto
// </auto-generated>
#pragma warning disable 0414, 1591
#region Designer generated code

using grpc = global::Grpc.Core;

namespace Pbx {
  /// <summary>
  /// This is the single method that needs to be implemented by a gRPC client.
  /// </summary>
  public static partial class Node
  {
    static readonly string __ServiceName = "pbx.Node";

    static readonly grpc::Marshaller<global::Pbx.ClientMsg> __Marshaller_pbx_ClientMsg = grpc::Marshallers.Create((arg) => global::Google.Protobuf.MessageExtensions.ToByteArray(arg), global::Pbx.ClientMsg.Parser.ParseFrom);
    static readonly grpc::Marshaller<global::Pbx.ServerMsg> __Marshaller_pbx_ServerMsg = grpc::Marshallers.Create((arg) => global::Google.Protobuf.MessageExtensions.ToByteArray(arg), global::Pbx.ServerMsg.Parser.ParseFrom);

    static readonly grpc::Method<global::Pbx.ClientMsg, global::Pbx.ServerMsg> __Method_MessageLoop = new grpc::Method<global::Pbx.ClientMsg, global::Pbx.ServerMsg>(
        grpc::MethodType.DuplexStreaming,
        __ServiceName,
        "MessageLoop",
        __Marshaller_pbx_ClientMsg,
        __Marshaller_pbx_ServerMsg);

    /// <summary>Service descriptor</summary>
    public static global::Google.Protobuf.Reflection.ServiceDescriptor Descriptor
    {
      get { return global::Pbx.ModelReflection.Descriptor.Services[0]; }
    }

    /// <summary>Base class for server-side implementations of Node</summary>
    public abstract partial class NodeBase
    {
      /// <summary>
      /// Client sends a stream of ClientMsg, server responds with a stream of ServerMsg
      /// </summary>
      /// <param name="requestStream">Used for reading requests from the client.</param>
      /// <param name="responseStream">Used for sending responses back to the client.</param>
      /// <param name="context">The context of the server-side call handler being invoked.</param>
      /// <returns>A task indicating completion of the handler.</returns>
      public virtual global::System.Threading.Tasks.Task MessageLoop(grpc::IAsyncStreamReader<global::Pbx.ClientMsg> requestStream, grpc::IServerStreamWriter<global::Pbx.ServerMsg> responseStream, grpc::ServerCallContext context)
      {
        throw new grpc::RpcException(new grpc::Status(grpc::StatusCode.Unimplemented, ""));
      }

    }

    /// <summary>Client for Node</summary>
    public partial class NodeClient : grpc::ClientBase<NodeClient>
    {
      /// <summary>Creates a new client for Node</summary>
      /// <param name="channel">The channel to use to make remote calls.</param>
      public NodeClient(grpc::Channel channel) : base(channel)
      {
      }
      /// <summary>Creates a new client for Node that uses a custom <c>CallInvoker</c>.</summary>
      /// <param name="callInvoker">The callInvoker to use to make remote calls.</param>
      public NodeClient(grpc::CallInvoker callInvoker) : base(callInvoker)
      {
      }
      /// <summary>Protected parameterless constructor to allow creation of test doubles.</summary>
      protected NodeClient() : base()
      {
      }
      /// <summary>Protected constructor to allow creation of configured clients.</summary>
      /// <param name="configuration">The client configuration.</param>
      protected NodeClient(ClientBaseConfiguration configuration) : base(configuration)
      {
      }

      /// <summary>
      /// Client sends a stream of ClientMsg, server responds with a stream of ServerMsg
      /// </summary>
      /// <param name="headers">The initial metadata to send with the call. This parameter is optional.</param>
      /// <param name="deadline">An optional deadline for the call. The call will be cancelled if deadline is hit.</param>
      /// <param name="cancellationToken">An optional token for canceling the call.</param>
      /// <returns>The call object.</returns>
      public virtual grpc::AsyncDuplexStreamingCall<global::Pbx.ClientMsg, global::Pbx.ServerMsg> MessageLoop(grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return MessageLoop(new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      /// <summary>
      /// Client sends a stream of ClientMsg, server responds with a stream of ServerMsg
      /// </summary>
      /// <param name="options">The options for the call.</param>
      /// <returns>The call object.</returns>
      public virtual grpc::AsyncDuplexStreamingCall<global::Pbx.ClientMsg, global::Pbx.ServerMsg> MessageLoop(grpc::CallOptions options)
      {
        return CallInvoker.AsyncDuplexStreamingCall(__Method_MessageLoop, null, options);
      }
      /// <summary>Creates a new instance of client from given <c>ClientBaseConfiguration</c>.</summary>
      protected override NodeClient NewInstance(ClientBaseConfiguration configuration)
      {
        return new NodeClient(configuration);
      }
    }

    /// <summary>Creates service definition that can be registered with a server</summary>
    /// <param name="serviceImpl">An object implementing the server-side handling logic.</param>
    public static grpc::ServerServiceDefinition BindService(NodeBase serviceImpl)
    {
      return grpc::ServerServiceDefinition.CreateBuilder()
          .AddMethod(__Method_MessageLoop, serviceImpl.MessageLoop).Build();
    }

  }
  /// <summary>
  /// Plugin interface.
  /// </summary>
  public static partial class Plugin
  {
    static readonly string __ServiceName = "pbx.Plugin";

    static readonly grpc::Marshaller<global::Pbx.ClientReq> __Marshaller_pbx_ClientReq = grpc::Marshallers.Create((arg) => global::Google.Protobuf.MessageExtensions.ToByteArray(arg), global::Pbx.ClientReq.Parser.ParseFrom);
    static readonly grpc::Marshaller<global::Pbx.ServerResp> __Marshaller_pbx_ServerResp = grpc::Marshallers.Create((arg) => global::Google.Protobuf.MessageExtensions.ToByteArray(arg), global::Pbx.ServerResp.Parser.ParseFrom);
    static readonly grpc::Marshaller<global::Pbx.SearchQuery> __Marshaller_pbx_SearchQuery = grpc::Marshallers.Create((arg) => global::Google.Protobuf.MessageExtensions.ToByteArray(arg), global::Pbx.SearchQuery.Parser.ParseFrom);
    static readonly grpc::Marshaller<global::Pbx.SearchFound> __Marshaller_pbx_SearchFound = grpc::Marshallers.Create((arg) => global::Google.Protobuf.MessageExtensions.ToByteArray(arg), global::Pbx.SearchFound.Parser.ParseFrom);
    static readonly grpc::Marshaller<global::Pbx.AccountEvent> __Marshaller_pbx_AccountEvent = grpc::Marshallers.Create((arg) => global::Google.Protobuf.MessageExtensions.ToByteArray(arg), global::Pbx.AccountEvent.Parser.ParseFrom);
    static readonly grpc::Marshaller<global::Pbx.Unused> __Marshaller_pbx_Unused = grpc::Marshallers.Create((arg) => global::Google.Protobuf.MessageExtensions.ToByteArray(arg), global::Pbx.Unused.Parser.ParseFrom);
    static readonly grpc::Marshaller<global::Pbx.TopicEvent> __Marshaller_pbx_TopicEvent = grpc::Marshallers.Create((arg) => global::Google.Protobuf.MessageExtensions.ToByteArray(arg), global::Pbx.TopicEvent.Parser.ParseFrom);
    static readonly grpc::Marshaller<global::Pbx.SubscriptionEvent> __Marshaller_pbx_SubscriptionEvent = grpc::Marshallers.Create((arg) => global::Google.Protobuf.MessageExtensions.ToByteArray(arg), global::Pbx.SubscriptionEvent.Parser.ParseFrom);
    static readonly grpc::Marshaller<global::Pbx.MessageEvent> __Marshaller_pbx_MessageEvent = grpc::Marshallers.Create((arg) => global::Google.Protobuf.MessageExtensions.ToByteArray(arg), global::Pbx.MessageEvent.Parser.ParseFrom);

    static readonly grpc::Method<global::Pbx.ClientReq, global::Pbx.ServerResp> __Method_FireHose = new grpc::Method<global::Pbx.ClientReq, global::Pbx.ServerResp>(
        grpc::MethodType.Unary,
        __ServiceName,
        "FireHose",
        __Marshaller_pbx_ClientReq,
        __Marshaller_pbx_ServerResp);

    static readonly grpc::Method<global::Pbx.SearchQuery, global::Pbx.SearchFound> __Method_Find = new grpc::Method<global::Pbx.SearchQuery, global::Pbx.SearchFound>(
        grpc::MethodType.Unary,
        __ServiceName,
        "Find",
        __Marshaller_pbx_SearchQuery,
        __Marshaller_pbx_SearchFound);

    static readonly grpc::Method<global::Pbx.AccountEvent, global::Pbx.Unused> __Method_Account = new grpc::Method<global::Pbx.AccountEvent, global::Pbx.Unused>(
        grpc::MethodType.Unary,
        __ServiceName,
        "Account",
        __Marshaller_pbx_AccountEvent,
        __Marshaller_pbx_Unused);

    static readonly grpc::Method<global::Pbx.TopicEvent, global::Pbx.Unused> __Method_Topic = new grpc::Method<global::Pbx.TopicEvent, global::Pbx.Unused>(
        grpc::MethodType.Unary,
        __ServiceName,
        "Topic",
        __Marshaller_pbx_TopicEvent,
        __Marshaller_pbx_Unused);

    static readonly grpc::Method<global::Pbx.SubscriptionEvent, global::Pbx.Unused> __Method_Subscription = new grpc::Method<global::Pbx.SubscriptionEvent, global::Pbx.Unused>(
        grpc::MethodType.Unary,
        __ServiceName,
        "Subscription",
        __Marshaller_pbx_SubscriptionEvent,
        __Marshaller_pbx_Unused);

    static readonly grpc::Method<global::Pbx.MessageEvent, global::Pbx.Unused> __Method_Message = new grpc::Method<global::Pbx.MessageEvent, global::Pbx.Unused>(
        grpc::MethodType.Unary,
        __ServiceName,
        "Message",
        __Marshaller_pbx_MessageEvent,
        __Marshaller_pbx_Unused);

    /// <summary>Service descriptor</summary>
    public static global::Google.Protobuf.Reflection.ServiceDescriptor Descriptor
    {
      get { return global::Pbx.ModelReflection.Descriptor.Services[1]; }
    }

    /// <summary>Base class for server-side implementations of Plugin</summary>
    public abstract partial class PluginBase
    {
      /// <summary>
      /// This plugin method is called by Tinode server for every message received from the clients. The 
      /// method returns a ServerCtrl message. Non-zero ServerCtrl.code indicates that no further 
      /// processing is needed. The Tinode server will generate a {ctrl} message from the returned ServerCtrl 
      /// and forward it to the client session. 
      /// ServerCtrl.code equals to 0 instructs the server to continue with default processing of the client message.
      /// </summary>
      /// <param name="request">The request received from the client.</param>
      /// <param name="context">The context of the server-side call handler being invoked.</param>
      /// <returns>The response to send back to the client (wrapped by a task).</returns>
      public virtual global::System.Threading.Tasks.Task<global::Pbx.ServerResp> FireHose(global::Pbx.ClientReq request, grpc::ServerCallContext context)
      {
        throw new grpc::RpcException(new grpc::Status(grpc::StatusCode.Unimplemented, ""));
      }

      /// <summary>
      /// An alteranative user and topic discovery mechanism. 
      /// A search request issued on a 'fnd' topic. This method is called to generate an alternative result set.
      /// </summary>
      /// <param name="request">The request received from the client.</param>
      /// <param name="context">The context of the server-side call handler being invoked.</param>
      /// <returns>The response to send back to the client (wrapped by a task).</returns>
      public virtual global::System.Threading.Tasks.Task<global::Pbx.SearchFound> Find(global::Pbx.SearchQuery request, grpc::ServerCallContext context)
      {
        throw new grpc::RpcException(new grpc::Status(grpc::StatusCode.Unimplemented, ""));
      }

      /// <summary>
      /// Account created, updated or deleted
      /// </summary>
      /// <param name="request">The request received from the client.</param>
      /// <param name="context">The context of the server-side call handler being invoked.</param>
      /// <returns>The response to send back to the client (wrapped by a task).</returns>
      public virtual global::System.Threading.Tasks.Task<global::Pbx.Unused> Account(global::Pbx.AccountEvent request, grpc::ServerCallContext context)
      {
        throw new grpc::RpcException(new grpc::Status(grpc::StatusCode.Unimplemented, ""));
      }

      /// <summary>
      /// Topic created, updated [or deleted -- not supported yet]
      /// </summary>
      /// <param name="request">The request received from the client.</param>
      /// <param name="context">The context of the server-side call handler being invoked.</param>
      /// <returns>The response to send back to the client (wrapped by a task).</returns>
      public virtual global::System.Threading.Tasks.Task<global::Pbx.Unused> Topic(global::Pbx.TopicEvent request, grpc::ServerCallContext context)
      {
        throw new grpc::RpcException(new grpc::Status(grpc::StatusCode.Unimplemented, ""));
      }

      /// <summary>
      /// Subscription created, updated or deleted
      /// </summary>
      /// <param name="request">The request received from the client.</param>
      /// <param name="context">The context of the server-side call handler being invoked.</param>
      /// <returns>The response to send back to the client (wrapped by a task).</returns>
      public virtual global::System.Threading.Tasks.Task<global::Pbx.Unused> Subscription(global::Pbx.SubscriptionEvent request, grpc::ServerCallContext context)
      {
        throw new grpc::RpcException(new grpc::Status(grpc::StatusCode.Unimplemented, ""));
      }

      /// <summary>
      /// Message published or deleted
      /// </summary>
      /// <param name="request">The request received from the client.</param>
      /// <param name="context">The context of the server-side call handler being invoked.</param>
      /// <returns>The response to send back to the client (wrapped by a task).</returns>
      public virtual global::System.Threading.Tasks.Task<global::Pbx.Unused> Message(global::Pbx.MessageEvent request, grpc::ServerCallContext context)
      {
        throw new grpc::RpcException(new grpc::Status(grpc::StatusCode.Unimplemented, ""));
      }

    }

    /// <summary>Client for Plugin</summary>
    public partial class PluginClient : grpc::ClientBase<PluginClient>
    {
      /// <summary>Creates a new client for Plugin</summary>
      /// <param name="channel">The channel to use to make remote calls.</param>
      public PluginClient(grpc::Channel channel) : base(channel)
      {
      }
      /// <summary>Creates a new client for Plugin that uses a custom <c>CallInvoker</c>.</summary>
      /// <param name="callInvoker">The callInvoker to use to make remote calls.</param>
      public PluginClient(grpc::CallInvoker callInvoker) : base(callInvoker)
      {
      }
      /// <summary>Protected parameterless constructor to allow creation of test doubles.</summary>
      protected PluginClient() : base()
      {
      }
      /// <summary>Protected constructor to allow creation of configured clients.</summary>
      /// <param name="configuration">The client configuration.</param>
      protected PluginClient(ClientBaseConfiguration configuration) : base(configuration)
      {
      }

      /// <summary>
      /// This plugin method is called by Tinode server for every message received from the clients. The 
      /// method returns a ServerCtrl message. Non-zero ServerCtrl.code indicates that no further 
      /// processing is needed. The Tinode server will generate a {ctrl} message from the returned ServerCtrl 
      /// and forward it to the client session. 
      /// ServerCtrl.code equals to 0 instructs the server to continue with default processing of the client message.
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="headers">The initial metadata to send with the call. This parameter is optional.</param>
      /// <param name="deadline">An optional deadline for the call. The call will be cancelled if deadline is hit.</param>
      /// <param name="cancellationToken">An optional token for canceling the call.</param>
      /// <returns>The response received from the server.</returns>
      public virtual global::Pbx.ServerResp FireHose(global::Pbx.ClientReq request, grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return FireHose(request, new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      /// <summary>
      /// This plugin method is called by Tinode server for every message received from the clients. The 
      /// method returns a ServerCtrl message. Non-zero ServerCtrl.code indicates that no further 
      /// processing is needed. The Tinode server will generate a {ctrl} message from the returned ServerCtrl 
      /// and forward it to the client session. 
      /// ServerCtrl.code equals to 0 instructs the server to continue with default processing of the client message.
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="options">The options for the call.</param>
      /// <returns>The response received from the server.</returns>
      public virtual global::Pbx.ServerResp FireHose(global::Pbx.ClientReq request, grpc::CallOptions options)
      {
        return CallInvoker.BlockingUnaryCall(__Method_FireHose, null, options, request);
      }
      /// <summary>
      /// This plugin method is called by Tinode server for every message received from the clients. The 
      /// method returns a ServerCtrl message. Non-zero ServerCtrl.code indicates that no further 
      /// processing is needed. The Tinode server will generate a {ctrl} message from the returned ServerCtrl 
      /// and forward it to the client session. 
      /// ServerCtrl.code equals to 0 instructs the server to continue with default processing of the client message.
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="headers">The initial metadata to send with the call. This parameter is optional.</param>
      /// <param name="deadline">An optional deadline for the call. The call will be cancelled if deadline is hit.</param>
      /// <param name="cancellationToken">An optional token for canceling the call.</param>
      /// <returns>The call object.</returns>
      public virtual grpc::AsyncUnaryCall<global::Pbx.ServerResp> FireHoseAsync(global::Pbx.ClientReq request, grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return FireHoseAsync(request, new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      /// <summary>
      /// This plugin method is called by Tinode server for every message received from the clients. The 
      /// method returns a ServerCtrl message. Non-zero ServerCtrl.code indicates that no further 
      /// processing is needed. The Tinode server will generate a {ctrl} message from the returned ServerCtrl 
      /// and forward it to the client session. 
      /// ServerCtrl.code equals to 0 instructs the server to continue with default processing of the client message.
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="options">The options for the call.</param>
      /// <returns>The call object.</returns>
      public virtual grpc::AsyncUnaryCall<global::Pbx.ServerResp> FireHoseAsync(global::Pbx.ClientReq request, grpc::CallOptions options)
      {
        return CallInvoker.AsyncUnaryCall(__Method_FireHose, null, options, request);
      }
      /// <summary>
      /// An alteranative user and topic discovery mechanism. 
      /// A search request issued on a 'fnd' topic. This method is called to generate an alternative result set.
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="headers">The initial metadata to send with the call. This parameter is optional.</param>
      /// <param name="deadline">An optional deadline for the call. The call will be cancelled if deadline is hit.</param>
      /// <param name="cancellationToken">An optional token for canceling the call.</param>
      /// <returns>The response received from the server.</returns>
      public virtual global::Pbx.SearchFound Find(global::Pbx.SearchQuery request, grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return Find(request, new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      /// <summary>
      /// An alteranative user and topic discovery mechanism. 
      /// A search request issued on a 'fnd' topic. This method is called to generate an alternative result set.
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="options">The options for the call.</param>
      /// <returns>The response received from the server.</returns>
      public virtual global::Pbx.SearchFound Find(global::Pbx.SearchQuery request, grpc::CallOptions options)
      {
        return CallInvoker.BlockingUnaryCall(__Method_Find, null, options, request);
      }
      /// <summary>
      /// An alteranative user and topic discovery mechanism. 
      /// A search request issued on a 'fnd' topic. This method is called to generate an alternative result set.
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="headers">The initial metadata to send with the call. This parameter is optional.</param>
      /// <param name="deadline">An optional deadline for the call. The call will be cancelled if deadline is hit.</param>
      /// <param name="cancellationToken">An optional token for canceling the call.</param>
      /// <returns>The call object.</returns>
      public virtual grpc::AsyncUnaryCall<global::Pbx.SearchFound> FindAsync(global::Pbx.SearchQuery request, grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return FindAsync(request, new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      /// <summary>
      /// An alteranative user and topic discovery mechanism. 
      /// A search request issued on a 'fnd' topic. This method is called to generate an alternative result set.
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="options">The options for the call.</param>
      /// <returns>The call object.</returns>
      public virtual grpc::AsyncUnaryCall<global::Pbx.SearchFound> FindAsync(global::Pbx.SearchQuery request, grpc::CallOptions options)
      {
        return CallInvoker.AsyncUnaryCall(__Method_Find, null, options, request);
      }
      /// <summary>
      /// Account created, updated or deleted
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="headers">The initial metadata to send with the call. This parameter is optional.</param>
      /// <param name="deadline">An optional deadline for the call. The call will be cancelled if deadline is hit.</param>
      /// <param name="cancellationToken">An optional token for canceling the call.</param>
      /// <returns>The response received from the server.</returns>
      public virtual global::Pbx.Unused Account(global::Pbx.AccountEvent request, grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return Account(request, new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      /// <summary>
      /// Account created, updated or deleted
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="options">The options for the call.</param>
      /// <returns>The response received from the server.</returns>
      public virtual global::Pbx.Unused Account(global::Pbx.AccountEvent request, grpc::CallOptions options)
      {
        return CallInvoker.BlockingUnaryCall(__Method_Account, null, options, request);
      }
      /// <summary>
      /// Account created, updated or deleted
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="headers">The initial metadata to send with the call. This parameter is optional.</param>
      /// <param name="deadline">An optional deadline for the call. The call will be cancelled if deadline is hit.</param>
      /// <param name="cancellationToken">An optional token for canceling the call.</param>
      /// <returns>The call object.</returns>
      public virtual grpc::AsyncUnaryCall<global::Pbx.Unused> AccountAsync(global::Pbx.AccountEvent request, grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return AccountAsync(request, new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      /// <summary>
      /// Account created, updated or deleted
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="options">The options for the call.</param>
      /// <returns>The call object.</returns>
      public virtual grpc::AsyncUnaryCall<global::Pbx.Unused> AccountAsync(global::Pbx.AccountEvent request, grpc::CallOptions options)
      {
        return CallInvoker.AsyncUnaryCall(__Method_Account, null, options, request);
      }
      /// <summary>
      /// Topic created, updated [or deleted -- not supported yet]
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="headers">The initial metadata to send with the call. This parameter is optional.</param>
      /// <param name="deadline">An optional deadline for the call. The call will be cancelled if deadline is hit.</param>
      /// <param name="cancellationToken">An optional token for canceling the call.</param>
      /// <returns>The response received from the server.</returns>
      public virtual global::Pbx.Unused Topic(global::Pbx.TopicEvent request, grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return Topic(request, new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      /// <summary>
      /// Topic created, updated [or deleted -- not supported yet]
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="options">The options for the call.</param>
      /// <returns>The response received from the server.</returns>
      public virtual global::Pbx.Unused Topic(global::Pbx.TopicEvent request, grpc::CallOptions options)
      {
        return CallInvoker.BlockingUnaryCall(__Method_Topic, null, options, request);
      }
      /// <summary>
      /// Topic created, updated [or deleted -- not supported yet]
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="headers">The initial metadata to send with the call. This parameter is optional.</param>
      /// <param name="deadline">An optional deadline for the call. The call will be cancelled if deadline is hit.</param>
      /// <param name="cancellationToken">An optional token for canceling the call.</param>
      /// <returns>The call object.</returns>
      public virtual grpc::AsyncUnaryCall<global::Pbx.Unused> TopicAsync(global::Pbx.TopicEvent request, grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return TopicAsync(request, new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      /// <summary>
      /// Topic created, updated [or deleted -- not supported yet]
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="options">The options for the call.</param>
      /// <returns>The call object.</returns>
      public virtual grpc::AsyncUnaryCall<global::Pbx.Unused> TopicAsync(global::Pbx.TopicEvent request, grpc::CallOptions options)
      {
        return CallInvoker.AsyncUnaryCall(__Method_Topic, null, options, request);
      }
      /// <summary>
      /// Subscription created, updated or deleted
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="headers">The initial metadata to send with the call. This parameter is optional.</param>
      /// <param name="deadline">An optional deadline for the call. The call will be cancelled if deadline is hit.</param>
      /// <param name="cancellationToken">An optional token for canceling the call.</param>
      /// <returns>The response received from the server.</returns>
      public virtual global::Pbx.Unused Subscription(global::Pbx.SubscriptionEvent request, grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return Subscription(request, new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      /// <summary>
      /// Subscription created, updated or deleted
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="options">The options for the call.</param>
      /// <returns>The response received from the server.</returns>
      public virtual global::Pbx.Unused Subscription(global::Pbx.SubscriptionEvent request, grpc::CallOptions options)
      {
        return CallInvoker.BlockingUnaryCall(__Method_Subscription, null, options, request);
      }
      /// <summary>
      /// Subscription created, updated or deleted
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="headers">The initial metadata to send with the call. This parameter is optional.</param>
      /// <param name="deadline">An optional deadline for the call. The call will be cancelled if deadline is hit.</param>
      /// <param name="cancellationToken">An optional token for canceling the call.</param>
      /// <returns>The call object.</returns>
      public virtual grpc::AsyncUnaryCall<global::Pbx.Unused> SubscriptionAsync(global::Pbx.SubscriptionEvent request, grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return SubscriptionAsync(request, new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      /// <summary>
      /// Subscription created, updated or deleted
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="options">The options for the call.</param>
      /// <returns>The call object.</returns>
      public virtual grpc::AsyncUnaryCall<global::Pbx.Unused> SubscriptionAsync(global::Pbx.SubscriptionEvent request, grpc::CallOptions options)
      {
        return CallInvoker.AsyncUnaryCall(__Method_Subscription, null, options, request);
      }
      /// <summary>
      /// Message published or deleted
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="headers">The initial metadata to send with the call. This parameter is optional.</param>
      /// <param name="deadline">An optional deadline for the call. The call will be cancelled if deadline is hit.</param>
      /// <param name="cancellationToken">An optional token for canceling the call.</param>
      /// <returns>The response received from the server.</returns>
      public virtual global::Pbx.Unused Message(global::Pbx.MessageEvent request, grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return Message(request, new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      /// <summary>
      /// Message published or deleted
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="options">The options for the call.</param>
      /// <returns>The response received from the server.</returns>
      public virtual global::Pbx.Unused Message(global::Pbx.MessageEvent request, grpc::CallOptions options)
      {
        return CallInvoker.BlockingUnaryCall(__Method_Message, null, options, request);
      }
      /// <summary>
      /// Message published or deleted
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="headers">The initial metadata to send with the call. This parameter is optional.</param>
      /// <param name="deadline">An optional deadline for the call. The call will be cancelled if deadline is hit.</param>
      /// <param name="cancellationToken">An optional token for canceling the call.</param>
      /// <returns>The call object.</returns>
      public virtual grpc::AsyncUnaryCall<global::Pbx.Unused> MessageAsync(global::Pbx.MessageEvent request, grpc::Metadata headers = null, global::System.DateTime? deadline = null, global::System.Threading.CancellationToken cancellationToken = default(global::System.Threading.CancellationToken))
      {
        return MessageAsync(request, new grpc::CallOptions(headers, deadline, cancellationToken));
      }
      /// <summary>
      /// Message published or deleted
      /// </summary>
      /// <param name="request">The request to send to the server.</param>
      /// <param name="options">The options for the call.</param>
      /// <returns>The call object.</returns>
      public virtual grpc::AsyncUnaryCall<global::Pbx.Unused> MessageAsync(global::Pbx.MessageEvent request, grpc::CallOptions options)
      {
        return CallInvoker.AsyncUnaryCall(__Method_Message, null, options, request);
      }
      /// <summary>Creates a new instance of client from given <c>ClientBaseConfiguration</c>.</summary>
      protected override PluginClient NewInstance(ClientBaseConfiguration configuration)
      {
        return new PluginClient(configuration);
      }
    }

    /// <summary>Creates service definition that can be registered with a server</summary>
    /// <param name="serviceImpl">An object implementing the server-side handling logic.</param>
    public static grpc::ServerServiceDefinition BindService(PluginBase serviceImpl)
    {
      return grpc::ServerServiceDefinition.CreateBuilder()
          .AddMethod(__Method_FireHose, serviceImpl.FireHose)
          .AddMethod(__Method_Find, serviceImpl.Find)
          .AddMethod(__Method_Account, serviceImpl.Account)
          .AddMethod(__Method_Topic, serviceImpl.Topic)
          .AddMethod(__Method_Subscription, serviceImpl.Subscription)
          .AddMethod(__Method_Message, serviceImpl.Message).Build();
    }

  }
}
#endregion
